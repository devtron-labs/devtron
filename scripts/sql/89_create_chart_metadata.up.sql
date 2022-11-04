-- Table Definition
CREATE TABLE "public"."chart_ref_metadata" (
    "chart_name" varchar(100) NOT NULL,
    "chart_description" varchar(300) NOT NULL
    PRIMARY KEY ("chart_name")
);