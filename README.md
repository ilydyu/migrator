## Install 
```go install github.com/ilydyu/migrator/cmd/@latest```

## Usage
migrator init - create directory, shema_history and schema_lock table

migrator create [migration name] - create migration. Example: migrator create create_users

migrator up - apply your migrations

migrator up dry - show what migration should be apply in the future

migrator down - rollback your last migration

migrator down step=[step] - rollback your migrations from end to start, where [step] - it is a number of migrations. Example: migrator down step=2

migrator history - show history
