-- integer y no smallint: mapea directo al int32 del contrato, sin conversiones.

-- Alumnos que participan en las batallas, con sus atributos físicos.
CREATE TABLE students (
  id uuid PRIMARY KEY,
  name text NOT NULL UNIQUE,
  strength integer NOT NULL CHECK (strength BETWEEN 0 AND 100),
  agility integer NOT NULL CHECK (agility BETWEEN 0 AND 100),
  stamina integer NOT NULL CHECK (stamina BETWEEN 0 AND 100)
);

-- Armas que los alumnos pueden usar en combate.
CREATE TABLE weapons (
  id uuid PRIMARY KEY,
  name text NOT NULL UNIQUE,
  damage integer NOT NULL CHECK (damage BETWEEN 0 AND 100),
  accuracy integer NOT NULL CHECK (accuracy BETWEEN 0 AND 100)
);

-- Localizaciones donde se desarrollan las batallas.
CREATE TABLE locations (
  id uuid PRIMARY KEY,
  name text NOT NULL UNIQUE,
  description text NOT NULL
);
