-- eportalgo/db/schema/07_logistics.sql

-- Library
CREATE TABLE library_books (
  book_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id  UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  title      TEXT NOT NULL,
  author     TEXT NOT NULL,
  isbn       TEXT,
  publisher  TEXT,
  category   TEXT,
  created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE book_copies (
  copy_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  book_id    UUID NOT NULL REFERENCES library_books(book_id) ON DELETE CASCADE,
  barcode    TEXT UNIQUE NOT NULL,
  status     TEXT NOT NULL DEFAULT 'Available',
  created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE library_loans (
  loan_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  copy_id     UUID NOT NULL REFERENCES book_copies(copy_id) ON DELETE CASCADE,
  user_id     UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  borrowed_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  due_date    DATE NOT NULL,
  returned_at TIMESTAMPTZ(6),
  status      TEXT NOT NULL DEFAULT 'Borrowed',
  fine_amount DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Inventory & Assets
CREATE TABLE inventory_items (
  item_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  item_name     TEXT NOT NULL,
  category      TEXT,
  quantity      INT NOT NULL DEFAULT 0,
  unit          TEXT,
  reorder_level INT,
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory_transactions (
  transaction_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id          UUID NOT NULL REFERENCES inventory_items(item_id) ON DELETE CASCADE,
  user_id          UUID REFERENCES users(user_id) ON DELETE SET NULL,
  transaction_type TEXT NOT NULL, -- IN, OUT
  quantity         INT NOT NULL,
  notes            TEXT,
  created_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE assets (
  asset_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id           UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  asset_name          TEXT NOT NULL,
  serial_number       TEXT,
  category            TEXT,
  purchase_date       DATE,
  value               DECIMAL(10, 2),
  status              TEXT NOT NULL DEFAULT 'Active',
  assigned_to_user_id UUID REFERENCES users(user_id) ON DELETE SET NULL,
  location            TEXT,
  created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Transport
CREATE TABLE transport_vehicles (
  vehicle_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id      UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  vehicle_number TEXT NOT NULL,
  capacity       INT NOT NULL,
  driver_name    TEXT,
  driver_phone   TEXT,
  created_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at     TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transport_routes (
  route_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id   UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  vehicle_id  UUID REFERENCES transport_vehicles(vehicle_id) ON DELETE SET NULL,
  route_name  TEXT NOT NULL,
  start_point TEXT,
  end_point   TEXT,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transport_stops (
  stop_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  route_id    UUID NOT NULL REFERENCES transport_routes(route_id) ON DELETE CASCADE,
  stop_name   TEXT NOT NULL,
  pickup_time TIME,
  drop_time   TIME,
  fare        DECIMAL(10, 2),
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transport_allocations (
  allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  user_id       UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  route_id      UUID NOT NULL REFERENCES transport_routes(route_id),
  stop_id       UUID NOT NULL REFERENCES transport_stops(stop_id),
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Hostel
CREATE TABLE hostel_buildings (
  building_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  building_name TEXT NOT NULL,
  type          TEXT NOT NULL, -- Boys, Girls, Mixed
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE hostel_rooms (
  room_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  building_id UUID NOT NULL REFERENCES hostel_buildings(building_id) ON DELETE CASCADE,
  room_number TEXT NOT NULL,
  capacity    INT NOT NULL,
  floor       INT,
  created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE hostel_allocations (
  allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  school_id     UUID NOT NULL REFERENCES schools(school_id) ON DELETE CASCADE,
  user_id       UUID UNIQUE NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  room_id       UUID NOT NULL REFERENCES hostel_rooms(room_id),
  start_date    DATE NOT NULL,
  end_date      DATE,
  status        TEXT NOT NULL DEFAULT 'Active',
  created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP
);
