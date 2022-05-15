CREATE TABLE IF NOT EXISTS "public"."self_registration_roles" (
     "role" varchar(255) NOT NULL,
     "created_on" timestamptz,
     "created_by" int4,
     "updated_on" timestamptz,
     "updated_by" int4
);