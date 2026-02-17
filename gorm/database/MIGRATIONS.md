# Database Migrations
This library uses a custom migration system built on GORM. Migrations are automatically applied when the application starts.

The migration system will:
- Connect to the database
- Create the `migration_records` table if needed
- Run any pending migrations
- Start the application

### Migration System Overview

- **Migration Files**: Located in `app/database/migrate/`
- **Naming Convention**: `XXX_description.go` (e.g., `001_create_users_table.go`)
- **Registration**: Migrations auto-register using `init()` functions
- **Tracking**: Applied migrations are tracked in the `migration_records` table
- **Automatic Execution**: Migrations run on application startup if database URL is configured
- **Automatic Rollback**: If any migration fails, all migrations applied in that session are automatically rolled back
- **Transaction Safety**: Each migration runs in its own transaction for atomicity

### How to Create New Migrations

1. Create a new migration file create a new migration directory, like `app/database/migrate/`, and
   put a new file like the below in it:

```go
// app/database/migrate/002_create_representatives_table.go
package migrate

import (
    "gorm.io/gorm"
    "github.com/Admiral-Piett/go-tools/gorm/database"
)

func init() {
    database.RegisterMigration(database.Migration{
        ID:          "002_create_representatives_table",
        Description: "Create representatives table for congressional members",
        Up: func(db *gorm.DB) error {
            return db.Exec(`
                CREATE TABLE representatives (
                    id SERIAL PRIMARY KEY,
                    name VARCHAR(255) NOT NULL,
                    state VARCHAR(2) NOT NULL,
                    district VARCHAR(10),
                    party VARCHAR(50) NOT NULL,
                    chamber VARCHAR(10) NOT NULL,
                    bioguide_id VARCHAR(10) UNIQUE NOT NULL,
                    active BOOLEAN DEFAULT TRUE,
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW()
                );
                
                CREATE INDEX idx_representatives_state ON representatives(state);
                CREATE INDEX idx_representatives_bioguide_id ON representatives(bioguide_id);
            `).Error
        },
        Down: func(db *gorm.DB) error {
            return db.Exec("DROP TABLE IF EXISTS representatives").Error
        },
    })
}
```

2. **Restart the application** - the migration will run automatically:

```bash
go run cmd/main.go
```

3. **Check the logs** to confirm the migration ran successfully:

```
INFO[0001] Running migration migration_id=002_create_representatives_table description="Create representatives table for congressional members"
INFO[0001] Database migrations completed applied_count=1
```

### Migration Best Practices

1. **Sequential numbering**: Use the next available number (001, 002, 003...)
2. **Descriptive names**: Use clear, descriptive migration IDs
3. **Include Down functions**: Always provide a way to rollback
4. **Use SQL DDL**: Write explicit CREATE/DROP statements for clarity
5. **Add indexes**: Include necessary indexes in the Up function
6. **Test thoroughly**: Verify both Up and Down functions work

### Automatic Rollback on Failure

The migration system includes automatic rollback functionality:

**How it works:**
- Each migration runs in its own transaction
- If any migration fails during startup, the system automatically rolls back ALL migrations applied in that session
- Rollback executes the `Down()` functions in reverse order
- Migration records are removed from the database
- The application startup fails with a clear error message

**Example scenario:**
```
INFO[0001] Running migration migration_id=001_create_users_table
INFO[0001] Running migration migration_id=002_create_representatives_table  
ERROR[0002] Migration failed, rolling back migration_id=002_create_representatives_table error="syntax error at position 45"
INFO[0002] Rolling back migration migration_id=002_create_representatives_table
INFO[0002] Rolling back migration migration_id=001_create_users_table
INFO[0002] Migration rollback completed
FATAL[0002] migration 002_create_representatives_table failed (successfully rolled back): syntax error
```

## Manual Migrations
You may however manually run the migrations or roll them back by a given ID.
**NOTE:** this may only be done for a single ID at a time.

```shell
go build -o app_name
app_name migrate <up|down> <id>
```

### Troubleshooting

**Migration fails**:
- Check database connection and permissions
- Verify PostgreSQL is running
- Review migration SQL syntax

**Reset development database**:
```bash
DROP DATABASE <database_name>;
CREATE DATABASE <database_name>;
```
Then restart the application to re-run all migrations.

**Check migration status**:
```sql
SELECT * FROM migration_records ORDER BY applied_at;
```

This migration system ensures consistent database schema evolution across all environments while keeping the implementation simple and maintainable.
