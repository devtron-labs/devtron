-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_git_sensor_node;

-- Create GitSensor node details table
CREATE TABLE git_sensor_node(
                                "id" INTEGER PRIMARY KEY DEFAULT nextval('id_seq_git_sensor_node'::regclass),
                                "host" VARCHAR NOT NULL,
                                "port" INTEGER NOT NULL,
                                "created_on" TIMESTAMPTZ,
                                "created_by" INTEGER,
                                "updated_on" TIMESTAMPTZ,
                                "updated_by" INTEGER
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_git_sensor_node_mapping;

-- Create GitSensorNode mapping table
CREATE TABLE git_sensor_node_mapping(
                                        "id" INTEGER PRIMARY KEY DEFAULT nextval('id_seq_git_sensor_node_mapping'::regclass),
                                        "app_id" INTEGER NOT NULL,
                                        "node_id" INTEGER NOT NULL,
                                        "created_on" TIMESTAMPTZ,
                                        "created_by" INTEGER,
                                        "updated_on" TIMESTAMPTZ,
                                        "updated_by" INTEGER
);