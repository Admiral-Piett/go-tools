package manual

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "github.com/Admiral-Piett/go-tools/gorm/database"
    log "github.com/sirupsen/logrus"
    "gorm.io/gorm"
    "os"
    "strings"
    "time"
)

func ProcessManualCommands(db *database.Database) error {
    firstCommand := strings.ToLower(os.Args[1])
    if firstCommand == "generate" {
        fmt.Println("Generating encryption encryptionKey")
        // Create a byte slice of length 32 for a 256-bit encryptionKey
        encryptionKey := make([]byte, 32)
        _, err := rand.Read(encryptionKey)
        if err != nil {
            log.Fatalf("Error generating random encryptionKey: %v", err)
        }
        fmt.Printf(
            "Generated Encryption Key (hex): %s\n",
            hex.EncodeToString(encryptionKey),
        )

        // For HMAC-SHA512, a 64-byte (512-bit) encryptionKey is recommended.
        secretKey := make([]byte, 64)
        _, err = rand.Read(secretKey)
        if err != nil {
            fmt.Printf("Error generating random encryptionKey: %v\n", err)
            return nil
        }
        fmt.Printf(
            "Generated HMAC Secret Key (hex): %s\n",
            hex.EncodeToString(secretKey),
        )
        return nil
    }
    if firstCommand != "migrate" {
        log.Warning(fmt.Sprintf("Invalid cli args: %s", os.Args))
        return nil
    }
    if len(os.Args) != 4 {
        log.Warning(fmt.Sprintf("Invalid cli args: %s", os.Args))
        return nil
    }

    var targetMigration *database.Migration
    for _, m := range database.MigrationRegistry {
        if os.Args[3] == m.Id {
            targetMigration = &m
            break
        }
    }
    if targetMigration == nil {
        log.Warning(fmt.Sprintf("Invalid migration id: %s", os.Args[3]))
        return nil
    }

    switch os.Args[2] {
    case "up":
        log.Info(
            fmt.Sprintf("Running manual migration: %s", targetMigration.Id),
        )
        err := db.Transaction(func(tx *gorm.DB) error {
            if err := targetMigration.Up(tx); err != nil {
                return err
            }

            // Record migration
            record := database.MigrationRecord{
                Id:          targetMigration.Id,
                Description: targetMigration.Description,
                AppliedAt:   time.Now().UTC(),
            }
            return tx.Create(&record).Error
        })
        if err != nil {
            log.WithError(err).Error("Manual Migration Failure")
            return nil
        }
    case "down":
        log.Info(
            fmt.Sprintf("Running manual rollback: %s", targetMigration.Id),
        )
        err := db.Transaction(func(tx *gorm.DB) error {
            if err := targetMigration.Down(tx); err != nil {
                return err
            }

            // Remove migration record
            return tx.Where("id = ?", targetMigration.Id).
                Delete(&database.MigrationRecord{}).
                Error
        })
        if err != nil {
            log.WithError(err).Error("Manual Rollback Failure")
            return nil
        }
    default:
        log.Warning(fmt.Sprintf("Invalid cli args: %s", os.Args))
        return nil
    }

    log.Info("Manual management complete, shutting down")
    return nil
}
