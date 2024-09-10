ALTER TABLE "public"."user_roles"
    ADD COLUMN "timeout_window_configuration_id" int;

ALTER TABLE "public"."user_roles" ADD FOREIGN KEY ("timeout_window_configuration_id")
    REFERENCES "public"."timeout_window_configuration" ("id");