CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource;

CREATE TABLE "public"."devtron_resource"
(
    "id"             int          NOT NULL DEFAULT nextval('id_seq_devtron_resource'::regclass),
    "kind"           varchar(250) NOT NULL,
    "display_name"   varchar(250) NOT NULL,
    "icon"           text,
    "parent_kind_id" int,
    "deleted"        boolean,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_schema;

CREATE TABLE "public"."devtron_resource_schema"
(
    "id"                  int        NOT NULL DEFAULT nextval('id_seq_devtron_resource_schema'::regclass),
    "devtron_resource_id" int,
    "version"             VARCHAR(5) NOT NULL,
    "schema"              jsonb,
    "latest"              boolean,
    "created_on"          timestamptz,
    "created_by"          integer,
    "updated_on"          timestamptz,
    "updated_by"          integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_object;

CREATE TABLE "public"."devtron_resource_object"
(
    "id"                         int          NOT NULL DEFAULT nextval('id_seq_devtron_resource_object'::regclass),
    "old_object_id"              int,
    "name"                       VARCHAR(250) NOT NULL,
    "devtron_resource_id"        int,
    "devtron_resource_schema_id" int,
    "object_data"                jsonb,
    "deleted"                    boolean,
    "created_on"                 timestamptz,
    "created_by"                 integer,
    "updated_on"                 timestamptz,
    "updated_by"                 integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_object_audit;

CREATE TABLE "public"."devtron_resource_object_audit"
(
    "id"                         int         NOT NULL DEFAULT nextval('id_seq_devtron_resource_object_audit'::regclass),
    "devtron_resource_object_id" int,
    "object_data"                json,
    "audit_operation"            VARCHAR(10) NOT NULL,
    "created_on"                 timestamptz,
    "created_by"                 integer,
    "updated_on"                 timestamptz,
    "updated_by"                 integer,
    PRIMARY KEY ("id")
);

INSERT INTO devtron_resource(kind,display_name,icon,parent_kind_id,deleted,created_on,created_by,updated_on,updated_by)
    VALUES('application','Applications','',0,false,now(),1,now(),1),
          ('devtron-application','Devtron applications','',(select id from devtron_resource where kind='application'),false,now(),1,now(),1),
          ('helm-application','Helm applications','',(select id from devtron_resource where kind='application'),false,now(),1,now(),1),
          ('job','Jobs','',0,false,now(),1,now(),1),
          ('cluster','Clusters','',0,false,now(),1,now(),1);

INSERT INTO devtron_resource_schema(devtron_resource_id,version,schema,latest,created_on,created_by,updated_on,updated_by)
VALUES((select id from devtron_resource where kind='devtron-application'),'v1',
       '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "Devtron Application Schema",
          "references": {
            "users": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                }
              },
              "required": [
                "id",
                "name"
              ]
            }
          },
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "icon": {
                  "type": "string",
                  "contentEncoding": "base64"
                },
                "Description": {
                  "type": "string"
                },
                "Created On": {
                  "type": "string",
                  "format": "date-time"
                },
                "Created By": {
                  "type": "object",
                  "refType": "#/references/users"
                },
                "Tags": {
                  "additionalProperties": {
                    "type": "string"
                  }
                },
                "metadata": {
                  "type": "object",
                  "properties": {
                    "Owners & Pager Duty": {
                      "type": "object",
                      "properties": {
                        "On Pager Duty": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        },
                        "Code Owners": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        }
                      },
                      "required": [
                        "Code Owners"
                      ]
                    },
                    "Service Details": {
                      "type": "object",
                      "properties": {
                        "Framework": {
                          "type": "string"
                        },
                        "Language": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        },
                        "Map": {
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "Communication Method": {
                          "type": "string"
                        },
                        "Internet Facing": {
                          "type": "boolean"
                        },
                        "Cities": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        }
                      }
                    },
                    "Documentation": {
                      "type": "object",
                      "properties": {
                        "Service Doc": {
                          "type": "string",
                          "format": "uri"
                        },
                        "API Contract": {
                          "type": "string",
                          "format": "uri"
                        },
                        "Runbook": {
                          "type": "string",
                          "format": "uri"
                        }
                      }
                    }
                  },
                  "required": [
                    "Owners & Pager Duty"
                  ]
                }
              },
              "required": [
                "id",
                "Created On",
                "Created By",
                "metadata"
              ]
            },
            "actions": {
              "type": "object"
            },
            "dependencies": {
              "type": "object"
            }
          },
          "required": [
            "version",
            "kind",
            "overview"
          ]
        }'
    ,true,now(),1,now(),1),
      ((select id from devtron_resource where kind='helm-application'),'v1',
       '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "Helm Application Schema",
          "references": {
            "users": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                }
              },
              "required": [
                "id",
                "name"
              ]
            }
          },
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "icon": {
                  "type": "string",
                  "contentEncoding": "base64"
                },
                "Description": {
                  "type": "string"
                },
                "Created On": {
                  "type": "string",
                  "format": "date-time"
                },
                "Created By": {
                  "type": "object",
                  "refType": "#/references/users"
                },
                "Tags": {
                  "additionalProperties": {
                    "type": "string"
                  }
                },
                "metadata": {
                  "type": "object",
                  "properties": {
                    "Owners & Pager Duty": {
                      "type": "object",
                      "properties": {
                        "On Pager Duty": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        },
                        "Code Owners": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        }
                      },
                      "required": [
                        "Code Owners"
                      ]
                    },
                    "Service Details": {
                      "type": "object",
                      "properties": {
                        "Framework": {
                          "type": "string"
                        },
                        "Language": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        },
                        "Map": {
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "Communication Method": {
                          "type": "string"
                        },
                        "Internet Facing": {
                          "type": "boolean"
                        },
                        "Cities": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        }
                      }
                    },
                    "Documentation": {
                      "type": "object",
                      "properties": {
                        "Service Doc": {
                          "type": "string",
                          "format": "uri"
                        },
                        "API Contract": {
                          "type": "string",
                          "format": "uri"
                        },
                        "Runbook": {
                          "type": "string",
                          "format": "uri"
                        }
                      }
                    }
                  },
                  "required": [
                    "Owners & Pager Duty"
                  ]
                }
              },
              "required": [
                "id",
                "Created On",
                "Created By",
                "metadata"
              ]
            },
            "actions": {
              "type": "object"
            },
            "dependencies": {
              "type": "object"
            }
          },
          "required": [
            "version",
            "kind",
            "overview"
          ]
        }',
       true,now(),1,now(),1),
      ((select id from devtron_resource where kind='job'),'v1',
       '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "Job Schema",
          "references": {
            "users": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                }
              },
              "required": [
                "id",
                "name"
              ]
            }
          },
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "icon": {
                  "type": "string",
                  "contentEncoding": "base64"
                },
                "Description": {
                  "type": "string"
                },
                "Created On": {
                  "type": "string",
                  "format": "date-time"
                },
                "Created By": {
                  "type": "object",
                  "refType": "#/references/users"
                },
                "Tags": {
                  "additionalProperties": {
                    "type": "string"
                  }
                },
                "metadata": {
                  "type": "object",
                  "properties": {
                    "Owners & Pager Duty": {
                      "type": "object",
                      "properties": {
                        "On Pager Duty": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        },
                        "Code Owners": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "refType": "#/references/users"
                          }
                        }
                      },
                      "required": [
                        "Code Owners"
                      ]
                    },
                    "Service Details": {
                      "type": "object",
                      "properties": {
                        "Framework": {
                          "type": "string"
                        },
                        "Language": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        },
                        "Map": {
                          "additionalProperties": {
                            "type": "string"
                          }
                        },
                        "Communication Method": {
                          "type": "string"
                        },
                        "Internet Facing": {
                          "type": "boolean"
                        },
                        "Cities": {
                          "type": "array",
                          "items": {
                            "type": "string"
                          }
                        }
                      }
                    },
                    "Documentation": {
                      "type": "object",
                      "properties": {
                        "Service Doc": {
                          "type": "string",
                          "format": "uri"
                        },
                        "API Contract": {
                          "type": "string",
                          "format": "uri"
                        },
                        "Runbook": {
                          "type": "string",
                          "format": "uri"
                        }
                      }
                    }
                  },
                  "required": [
                    "Owners & Pager Duty"
                  ]
                }
              },
              "required": [
                "id",
                "Created On",
                "Created By",
                "metadata"
              ]
            },
            "actions": {
              "type": "object"
            },
            "dependencies": {
              "type": "object"
            }
          },
          "required": [
            "version",
            "kind",
            "overview"
          ]
        }',
       true,now(),1,now(),1),
      ((select id from devtron_resource where kind='cluster'),'v1',
       '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "Cluster Schema",
          "references": {
            "users": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                }
              },
              "required": [
                "id",
                "name"
              ]
            }
          },
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string"
                },
                "name": {
                  "type": "string"
                },
                "icon": {
                  "type": "string",
                  "contentEncoding": "base64"
                },
                "Description": {
                  "type": "string"
                },
                "Created On": {
                  "type": "string",
                  "format": "date-time"
                },
                "Created By": {
                  "type": "object",
                  "refType": "#/references/users"
                },
                "Tags": {
                  "additionalProperties": {
                    "type": "string"
                  }
                },
                "metadata": {
                  "type": "object",
                  "properties": {
                    "POCs": {
                      "type": "string"
                    },
                    "K8s Version": {
                      "type": "string"
                    },
                    "Cluster Provider": {
                      "type": "string"
                    }
                  }
                }
              },
              "required": [
                "id",
                "Created On",
                "Created By",
                "metadata"
              ]
            },
            "actions": {
              "type": "object"
            },
            "dependencies": {
              "type": "object"
            }
          },
          "required": [
            "version",
            "kind",
            "overview"
          ]
        }',
       true,now(),1,now(),1),;
