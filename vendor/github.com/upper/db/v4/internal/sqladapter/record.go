package sqladapter

import (
	"reflect"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

func recordID(store db.Store, record db.Record) (db.Cond, error) {
	if record == nil {
		return nil, db.ErrNilRecord
	}

	if hasConstraints, ok := record.(db.HasConstraints); ok {
		return hasConstraints.Constraints(), nil
	}

	id := db.Cond{}

	keys, fields, err := recordPrimaryKeyFieldValues(store, record)
	if err != nil {
		return nil, err
	}
	for i := range fields {
		if fields[i] == reflect.Zero(reflect.TypeOf(fields[i])).Interface() {
			return nil, db.ErrRecordIDIsZero
		}
		id[keys[i]] = fields[i]
	}
	if len(id) < 1 {
		return nil, db.ErrRecordIDIsZero
	}

	return id, nil
}

func recordPrimaryKeyFieldValues(store db.Store, record db.Record) ([]string, []interface{}, error) {
	sess := store.Session()

	pKeys, err := sess.(Session).PrimaryKeys(store.Name())
	if err != nil {
		return nil, nil, err
	}

	fields := sqlbuilder.Mapper.FieldsByName(reflect.ValueOf(record), pKeys)

	values := make([]interface{}, 0, len(fields))
	for i := range fields {
		if fields[i].IsValid() {
			values = append(values, fields[i].Interface())
		}
	}

	return pKeys, values, nil
}

func recordCreate(store db.Store, record db.Record) error {
	sess := store.Session()

	if validator, ok := record.(db.Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	if hook, ok := record.(db.BeforeCreateHook); ok {
		if err := hook.BeforeCreate(sess); err != nil {
			return err
		}
	}

	if creator, ok := store.(db.StoreCreator); ok {
		if err := creator.Create(record); err != nil {
			return err
		}
	} else {
		if err := store.InsertReturning(record); err != nil {
			return err
		}
	}

	if hook, ok := record.(db.AfterCreateHook); ok {
		if err := hook.AfterCreate(sess); err != nil {
			return err
		}
	}
	return nil
}

func recordUpdate(store db.Store, record db.Record) error {
	sess := store.Session()

	if validator, ok := record.(db.Validator); ok {
		if err := validator.Validate(); err != nil {
			return err
		}
	}

	if hook, ok := record.(db.BeforeUpdateHook); ok {
		if err := hook.BeforeUpdate(sess); err != nil {
			return err
		}
	}

	if updater, ok := store.(db.StoreUpdater); ok {
		if err := updater.Update(record); err != nil {
			return err
		}
	} else {
		if err := record.Store(sess).UpdateReturning(record); err != nil {
			return err
		}
	}

	if hook, ok := record.(db.AfterUpdateHook); ok {
		if err := hook.AfterUpdate(sess); err != nil {
			return err
		}
	}
	return nil
}
