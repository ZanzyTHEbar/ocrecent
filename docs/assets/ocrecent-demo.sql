CREATE TABLE session (
  id TEXT PRIMARY KEY,
  parent_id TEXT,
  title TEXT,
  time_created INTEGER,
  time_updated INTEGER,
  time_archived INTEGER,
  directory TEXT,
  path TEXT,
  project_id TEXT
);
CREATE TABLE project (
  id TEXT PRIMARY KEY,
  name TEXT,
  worktree TEXT
);

INSERT INTO project VALUES ('p_code', 'ocrecent', '/home/demo/code');
INSERT INTO project VALUES ('p_docs', 'docs', '/home/demo/docs');
INSERT INTO session VALUES ('ses_demo_build', NULL, 'Fix release script', 4102444800000, 4102444800000, 0, '/home/demo/code', '', 'p_code');
INSERT INTO session VALUES ('ses_demo_docs', NULL, 'Document CLI demo', 4102444700000, 4102444700000, 0, '/home/demo/docs', '', 'p_docs');
INSERT INTO session VALUES ('ses_demo_parser', NULL, 'Refactor parser', 4102444600000, 4102444600000, 0, '/home/demo/code/parser', 'parser', 'p_code');
