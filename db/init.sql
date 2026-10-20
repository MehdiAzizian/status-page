CREATE TABLE components (
  id     SERIAL PRIMARY KEY,
  name   TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'operational'
);
INSERT INTO components (name) VALUES ('Website'), ('Login'), ('API');
