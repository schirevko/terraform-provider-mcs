---
layout: "mcs"
page_title: "mcs: db_database"
subcategory: ""
description: |-
  Manages a db database.
---

# mcs\_db\_database (Resource)

Provides a db database resource. This can be used to create, modify and delete db databases.

## Example Usage

```terraform

resource "mcs_db_database" "mydb" {
  name        = "mydb"
  instance_id = example_db_instance_id
  charset     = "utf8"
  collate     = "utf8_general_ci"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the database. Changing this creates a new database.

* `instance_id` - (Optional) ID of the instance that database is created for. **Deprecated** Please, use `dbms_id` attribute instead.

* `dbms_id` - (Optional) ID of the instance or cluster that database is created for.

* `charset` - Type of charset used for the database. Changing this creates a new database.

* `collate` - Collate option of the database.  Changing this creates a new database.

Either `instance_id` or `dbms_id` must be configured.
