package database

// DataOperations defines a basic generic interface for handling data models in the database.
// It standardizes the Add, Update, and Delete operations.
// This interface can be implemented by database handlers to provide a consistent API.
//
// T is the type of the data model (e.g., *Tool).
type DataOperations[T any] interface {
	Get(id int64) (T, error)
	List() ([]T, error)
	// Add creates a new record for the given model.
	// It may take a user for auditing purposes and may return the ID of the new record.
	Add(model T, user *User) (int64, error)

	// Update modifies an existing record.
	// It may take a user for auditing purposes. The model should contain its ID.
	Update(model T, user *User) error

	// Delete removes a record from the database by its ID.
	// It may take a user for auditing purposes.
	Delete(id int64, user *User) error
}
