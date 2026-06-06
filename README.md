## Install 
```go install github.com/ilydyu/migrator/cmd/migrator@latest```

## Usage

First, create directory struct and config ```migrator setup```

Second, need create tables for migrator ```migrator init```

Third, create your migration ```migrator create create_users```

Four, write sql and apply your migration ```migrator up```

You are awesome


## Help

migrator setup - create directory and config file

migrator init - create shema_history and schema_lock table in database

migrator create [migration name] - create migration. Example: migrator create create_users

migrator up - apply your migrations

migrator up dry - show what migration should be apply in the future

migrator down - rollback your last migration

migrator down step=[step] - rollback your migrations from end to start, where [step] - it is a number of migrations. Example: migrator down step=2

migrator history - show history
