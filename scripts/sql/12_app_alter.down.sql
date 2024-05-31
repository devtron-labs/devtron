/*
 * Copyright (c) 2024. Devtron Inc.
 */

CREATE UNIQUE INDEX "app_app_name_key" ON "public"."app" USING BTREE ("app_name");