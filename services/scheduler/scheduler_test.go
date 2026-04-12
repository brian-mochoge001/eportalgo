package scheduler

import (
	"testing"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

func TestFitnessCalculation(t *testing.T) {
	schoolID := uuid.New()
	classID := uuid.New()
	subjectID := uuid.New()
	teacherID := uuid.New()
	roomID := uuid.New()

	data := &InputData{
		Classes: []db.GetClassesForSchedulingRow{
			{ClassID: classID, SchoolID: schoolID, ClassName: "Class A", EnrollmentCount: 30},
		},
		Subjects: []db.Subject{
			{SubjectID: subjectID, SchoolID: schoolID, SubjectName: "Math"},
		},
		Teachers: []db.GetTeachersBySchoolRow{
			{UserID: teacherID, FirstName: "John", LastName: "Doe"},
		},
		Rooms: []db.GetRoomsBySchoolRow{
			{RoomID: roomID, SchoolID: schoolID, RoomName: "Room 101", Capacity: 40},
		},
		TeacherSubs:  map[uuid.UUID][]uuid.UUID{teacherID: {subjectID}},
		CourseSubs:   map[uuid.UUID][]uuid.UUID{uuid.New(): {subjectID}},
		Availability: make(map[uuid.UUID][]db.TeacherAvailability),
	}

	s := NewScheduler(nil, Config{})

	// Case 1: Perfect timetable
	c1 := Chromosome{
		Genes: []Gene{
			{
				ClassID:   classID,
				SubjectID: subjectID,
				TeacherID: teacherID,
				RoomID:    roomID,
				DayOfWeek: 1,
				StartTime: time.Date(0, 0, 0, 8, 0, 0, 0, time.UTC),
				EndTime:   time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC),
				Duration:  time.Hour,
			},
		},
	}
	f1 := s.calculateFitness(c1, data)
	if f1 < 0.9 {
		t.Errorf("Expected high fitness for perfect timetable, got %f", f1)
	}

	// Case 2: Room capacity conflict
	data.Rooms[0].Capacity = 20 // Enrollment is 30
	f2 := s.calculateFitness(c1, data)
	if f2 >= f1 {
		t.Errorf("Expected lower fitness for capacity conflict, got %f >= %f", f2, f1)
	}
}
