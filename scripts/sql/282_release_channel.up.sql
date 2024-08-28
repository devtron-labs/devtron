INSERT INTO devtron_resource(kind, display_name, icon,is_exposed, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('release-channel', 'Release Channel', '',true, 0, false, now(), 1, now(), 1);

INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, sample_schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'release-channel'), 'alpha1',
        '{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Release Channel Schema",
  "type": "object",
  "properties":
  {
    "kind":
    {
      "const": "release-channel"
    },
    "version":
    {
      "type": "string",
      "enum":
      [
        "alpha1"
      ]
    },
    "overview":
    {
      "type": "object",
      "properties":
      {
        "id":
        {
          "type": "number"
        },
        "idType":
        {
          "type": "string",
          "enum":
          [
            "resourceObjectId",
            "oldObjectId"
          ]
        },
        "name":
        {
          "type": "string"
        },
        "icon":
        {
          "type": "string",
          "format": "uri"
        },
        "description":
        {
          "type": "string"
        },
        "releaseChannelId":
        {
          "type": "string"
        },
        "createdOn":
        {
          "type": "string"
        },
        "createdBy":
        {
          "type": "object",
          "refType": "#/references/users"
        },
        "default":
        {
         "type": "boolean"
        },
        "tags":
        {
          "additionalProperties":
          {
            "type": "string"
          }
        },
        "metadata":
        {
          "type": "object",
          "properties":
          {}
        },
        "required":
        [
          "id",
          "idType",
          "releaseChannelId"
        ]
      }
    },
    "dependencies":
    {
      "type": "array"
    }
  },
  "required":
  [
    "version",
    "kind",
    "overview"
  ]
}','{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Release Channel Schema",
  "type": "object",
  "properties":
  {
    "kind":
    {
      "const": "release-channel"
    },
    "version":
    {
      "type": "string",
      "enum":
      [
        "alpha1"
      ]
    },
    "overview":
    {
      "type": "object",
      "properties":
      {
        "id":
        {
          "type": "number"
        },
        "idType":
        {
          "type": "string",
          "enum":
          [
            "resourceObjectId",
            "oldObjectId"
          ]
        },
        "name":
        {
          "type": "string"
        },
        "icon":
        {
          "type": "string",
          "format": "uri"
        },
        "description":
        {
          "type": "string"
        },
        "releaseChannelId":
        {
          "type": "string"
        },
        "createdOn":
        {
          "type": "string"
        },
        "createdBy":
        {
          "type": "object",
          "refType": "#/references/users"
        },
        "default":
        {
         "type": "boolean"
        },
        "tags":
        {
          "additionalProperties":
          {
            "type": "string"
          }
        },
        "metadata":
        {
          "type": "object",
          "properties":
          {}
        },
        "required":
        [
          "id",
          "idType",
          "releaseChannelId"
        ]
      }
    },
    "dependencies":
    {
      "type": "array"
    }
  },
  "required":
  [
    "version",
    "kind",
    "overview"
  ]
}',true, now(), 1, now(), 1);


