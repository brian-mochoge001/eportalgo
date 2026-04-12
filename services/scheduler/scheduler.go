package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// Gene represents a single class meeting
type Gene struct {
	ClassID   uuid.UUID
	SubjectID uuid.UUID
	TeacherID uuid.UUID
	RoomID    uuid.UUID
	DayOfWeek int
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// Chromosome represents a complete timetable
type Chromosome struct {
	Genes   []Gene
	Fitness float64
}

// Scheduler handles the genetic algorithm
type Scheduler struct {
	Queries *db.Queries
	Config  Config
}

type Config struct {
	PopulationSize     int
	MaxGenerations     int
	MutationRate       float64
	CrossoverRate      float64
	SchoolDayStart     int // Hour (e.g., 8)
	SchoolDayEnd       int // Hour (e.g., 17)
	MaxConsecutive     int // Max consecutive classes for teachers
	BreakRequiredAfter int // Number of classes before a break is required
}

func NewScheduler(q *db.Queries, cfg Config) *Scheduler {
	if cfg.PopulationSize == 0 {
		cfg.PopulationSize = 100
	}
	if cfg.MaxGenerations == 0 {
		cfg.MaxGenerations = 500
	}
	if cfg.MutationRate == 0 {
		cfg.MutationRate = 0.05
	}
	if cfg.CrossoverRate == 0 {
		cfg.CrossoverRate = 0.8
	}
	if cfg.SchoolDayStart == 0 {
		cfg.SchoolDayStart = 8
	}
	if cfg.SchoolDayEnd == 0 {
		cfg.SchoolDayEnd = 17
	}
	if cfg.MaxConsecutive == 0 {
		cfg.MaxConsecutive = 3
	}
	return &Scheduler{Queries: q, Config: cfg}
}

// Data needed for scheduling
type InputData struct {
	Classes      []db.GetClassesForSchedulingRow
	Subjects     []db.Subject
	Teachers     []db.GetTeachersBySchoolRow
	Rooms        []db.GetRoomsBySchoolRow
	TeacherSubs  map[uuid.UUID][]uuid.UUID // TeacherID -> []SubjectID
	CourseSubs   map[uuid.UUID][]uuid.UUID // CourseID -> []SubjectID
	Availability map[uuid.UUID][]db.TeacherAvailability
}

func (s *Scheduler) Generate(ctx context.Context, schoolID uuid.UUID, academicYear, semester string) (*Chromosome, error) {
	data, err := s.fetchInputData(ctx, schoolID, academicYear, semester)
	if err != nil {
		return nil, err
	}

	if len(data.Classes) == 0 {
		return nil, fmt.Errorf("no classes found for scheduling")
	}
	if len(data.Rooms) == 0 {
		return nil, fmt.Errorf("no rooms found for scheduling")
	}

	population := s.initializePopulation(data)

	for gen := 0; gen < s.Config.MaxGenerations; gen++ {
		for i := range population {
			population[i].Fitness = s.calculateFitness(population[i], data)
		}

		sort.Slice(population, func(i, j int) bool {
			return population[i].Fitness > population[j].Fitness
		})

		if population[0].Fitness >= 1.0 {
			break
		}

		newPopulation := make([]Chromosome, 0, s.Config.PopulationSize)
		newPopulation = append(newPopulation, population[0])

		for len(newPopulation) < s.Config.PopulationSize {
			parent1 := s.selectParent(population)
			parent2 := s.selectParent(population)
			child1, child2 := s.crossover(parent1, parent2)
			s.mutate(&child1, data)
			s.mutate(&child2, data)
			newPopulation = append(newPopulation, child1)
			if len(newPopulation) < s.Config.PopulationSize {
				newPopulation = append(newPopulation, child2)
			}
		}
		population = newPopulation
	}

	return &population[0], nil
}

func (s *Scheduler) fetchInputData(ctx context.Context, schoolID uuid.UUID, academicYear, semester string) (*InputData, error) {
	classes, err := s.Queries.GetClassesForScheduling(ctx, db.GetClassesForSchedulingParams{
		SchoolID:     schoolID,
		AcademicYear: academicYear,
		Semester:     toNullString(semester),
	})
	if err != nil {
		return nil, err
	}

	subjects, err := s.Queries.GetSubjectsBySchool(ctx, schoolID)
	if err != nil {
		return nil, err
	}

	teachers, err := s.Queries.GetTeachersBySchool(ctx, schoolID)
	if err != nil {
		return nil, err
	}

	rooms, err := s.Queries.GetRoomsBySchool(ctx, schoolID)
	if err != nil {
		return nil, err
	}

	teacherSubsRaw, err := s.Queries.ListTeacherSubjectsBySchool(ctx, schoolID)
	if err != nil {
		return nil, err
	}
	teacherSubs := make(map[uuid.UUID][]uuid.UUID)
	for _, ts := range teacherSubsRaw {
		teacherSubs[ts.TeacherID] = append(teacherSubs[ts.TeacherID], ts.SubjectID)
	}

	courseSubs := make(map[uuid.UUID][]uuid.UUID)
	for _, class := range classes {
		subs, err := s.Queries.GetCourseSubjects(ctx, class.CourseID)
		if err == nil {
			ids := make([]uuid.UUID, len(subs))
			for i, sub := range subs {
				ids[i] = sub.SubjectID
			}
			courseSubs[class.CourseID] = ids
		}
	}

	availRaw, err := s.Queries.GetTeacherAvailabilities(ctx, uuid.NullUUID{})
	if err != nil {
		return nil, err
	}
	availability := make(map[uuid.UUID][]db.TeacherAvailability)
	for _, a := range availRaw {
		availability[a.TeacherID] = append(availability[a.TeacherID], a)
	}

	return &InputData{
		Classes:      classes,
		Subjects:     subjects,
		Teachers:     teachers,
		Rooms:        rooms,
		TeacherSubs:  teacherSubs,
		CourseSubs:   courseSubs,
		Availability: availability,
	}, nil
}

func (s *Scheduler) initializePopulation(data *InputData) []Chromosome {
	population := make([]Chromosome, s.Config.PopulationSize)
	for i := 0; i < s.Config.PopulationSize; i++ {
		population[i] = s.generateRandomChromosome(data)
	}
	return population
}

func (s *Scheduler) generateRandomChromosome(data *InputData) Chromosome {
	genes := make([]Gene, 0)
	for _, class := range data.Classes {
		subs := data.CourseSubs[class.CourseID]
		for _, subID := range subs {
			var subject db.Subject
			for _, sub := range data.Subjects {
				if sub.SubjectID == subID {
					subject = sub
					break
				}
			}

			qualifiedTeachers := make([]uuid.UUID, 0)
			for tID, ts := range data.TeacherSubs {
				for _, sid := range ts {
					if sid == subID {
						qualifiedTeachers = append(qualifiedTeachers, tID)
						break
					}
				}
			}
			if len(qualifiedTeachers) == 0 {
				continue
			}
			teacherID := qualifiedTeachers[rand.Intn(len(qualifiedTeachers))]

			roomID := data.Rooms[rand.Intn(len(data.Rooms))].RoomID

			duration := time.Hour
			if subject.LabPeriodRequired {
				duration = 2 * time.Hour
			} else if subject.DoublePeriodRequired {
				duration = 2 * time.Hour
			}

			day := rand.Intn(5) + 1
			maxHour := s.Config.SchoolDayEnd - int(duration.Hours())
			if maxHour <= s.Config.SchoolDayStart {
				maxHour = s.Config.SchoolDayStart + 1
			}
			hour := rand.Intn(maxHour-s.Config.SchoolDayStart) + s.Config.SchoolDayStart
			startTime := time.Date(0, 0, 0, hour, 0, 0, 0, time.UTC)
			endTime := startTime.Add(duration)

			genes = append(genes, Gene{
				ClassID:   class.ClassID,
				SubjectID: subID,
				TeacherID: teacherID,
				RoomID:    roomID,
				DayOfWeek: day,
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  duration,
			})
		}
	}
	return Chromosome{Genes: genes}
}

func (s *Scheduler) calculateFitness(c Chromosome, data *InputData) float64 {
	score := 2000.0
	hardConflicts := 0
	softConflicts := 0

	// Track teacher schedules for consecutive classes
	teacherSchedules := make(map[uuid.UUID]map[int][]Gene)

	for i, g1 := range c.Genes {
		// Room Capacity Check
		var room db.GetRoomsBySchoolRow
		for _, r := range data.Rooms {
			if r.RoomID == g1.RoomID {
				room = r
				break
			}
		}
		var class db.GetClassesForSchedulingRow
		for _, cl := range data.Classes {
			if cl.ClassID == g1.ClassID {
				class = cl
				break
			}
		}
		if int32(class.EnrollmentCount) > room.Capacity {
			hardConflicts++
		}

		// Overlap checks
		for j, g2 := range c.Genes {
			if i == j {
				continue
			}
			if g1.DayOfWeek == g2.DayOfWeek {
				overlap := g1.StartTime.Before(g2.EndTime) && g2.StartTime.Before(g1.EndTime)
				if overlap {
					if g1.TeacherID == g2.TeacherID {
						hardConflicts++
					}
					if g1.RoomID == g2.RoomID {
						hardConflicts++
					}
					if g1.ClassID == g2.ClassID {
						hardConflicts++
					}
				}
			}
		}

		// Teacher availability
		teacherAvails := data.Availability[g1.TeacherID]
		if len(teacherAvails) > 0 {
			available := false
			for _, a := range teacherAvails {
				if a.DayOfWeek == int32(g1.DayOfWeek) {
					aStart := time.Date(0, 0, 0, a.StartTime.Hour(), a.StartTime.Minute(), 0, 0, time.UTC)
					aEnd := time.Date(0, 0, 0, a.EndTime.Hour(), a.EndTime.Minute(), 0, 0, time.UTC)
					if !g1.StartTime.Before(aStart) && !g1.EndTime.After(aEnd) {
						available = true
						break
					}
				}
			}
			if !available {
				hardConflicts++
			}
		}

		// Collect teacher schedule
		if teacherSchedules[g1.TeacherID] == nil {
			teacherSchedules[g1.TeacherID] = make(map[int][]Gene)
		}
		teacherSchedules[g1.TeacherID][g1.DayOfWeek] = append(teacherSchedules[g1.TeacherID][g1.DayOfWeek], g1)
	}

	// Soft Constraints: Consecutive classes and breaks
	for _, days := range teacherSchedules {
		for _, dayGenes := range days {
			sort.Slice(dayGenes, func(i, j int) bool {
				return dayGenes[i].StartTime.Before(dayGenes[j].StartTime)
			})

			consecutive := 0
			for i := 0; i < len(dayGenes); i++ {
				if i > 0 {
					// Check if gap is small (consecutive)
					gap := dayGenes[i].StartTime.Sub(dayGenes[i-1].EndTime)
					if gap <= 15*time.Minute {
						consecutive++
					} else {
						consecutive = 0
					}
				}
				if consecutive >= s.Config.MaxConsecutive {
					softConflicts++
				}
			}

			// Check for lunch break (e.g., between 12 and 14)
			hasBreak := false
			for hour := 12; hour < 14; hour++ {
				breakTime := time.Date(0, 0, 0, hour, 0, 0, 0, time.UTC)
				busy := false
				for _, g := range dayGenes {
					if g.StartTime.Before(breakTime.Add(time.Hour)) && breakTime.Before(g.EndTime) {
						busy = true
						break
					}
				}
				if !busy {
					hasBreak = true
					break
				}
			}
			if !hasBreak {
				softConflicts++
			}
		}
	}

	score -= float64(hardConflicts) * 100.0
	score -= float64(softConflicts) * 10.0
	if score < 0 {
		score = 0
	}

	return score / 2000.0
}

func (s *Scheduler) selectParent(population []Chromosome) Chromosome {
	tournamentSize := 5
	best := population[rand.Intn(len(population))]
	for i := 0; i < tournamentSize-1; i++ {
		competitor := population[rand.Intn(len(population))]
		if competitor.Fitness > best.Fitness {
			best = competitor
		}
	}
	return best
}

func (s *Scheduler) crossover(p1, p2 Chromosome) (Chromosome, Chromosome) {
	if rand.Float64() > s.Config.CrossoverRate || len(p1.Genes) == 0 {
		return p1, p2
	}
	point := rand.Intn(len(p1.Genes))
	c1Genes := append([]Gene{}, p1.Genes[:point]...)
	c1Genes = append(c1Genes, p2.Genes[point:]...)
	c2Genes := append([]Gene{}, p2.Genes[:point]...)
	c2Genes = append(c2Genes, p1.Genes[point:]...)
	return Chromosome{Genes: c1Genes}, Chromosome{Genes: c2Genes}
}

func (s *Scheduler) mutate(c *Chromosome, data *InputData) {
	for i := range c.Genes {
		if rand.Float64() < s.Config.MutationRate {
			if rand.Float64() < 0.5 {
				c.Genes[i].DayOfWeek = rand.Intn(5) + 1
				maxHour := s.Config.SchoolDayEnd - int(c.Genes[i].Duration.Hours())
				if maxHour <= s.Config.SchoolDayStart {
					maxHour = s.Config.SchoolDayStart + 1
				}
				hour := rand.Intn(maxHour-s.Config.SchoolDayStart) + s.Config.SchoolDayStart
				c.Genes[i].StartTime = time.Date(0, 0, 0, hour, 0, 0, 0, time.UTC)
				c.Genes[i].EndTime = c.Genes[i].StartTime.Add(c.Genes[i].Duration)
			} else {
				c.Genes[i].RoomID = data.Rooms[rand.Intn(len(data.Rooms))].RoomID
			}
		}
	}
}
