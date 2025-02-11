package sqladapter

import (
	"fmt"
	"reflect"

	db "github.com/upper/db/v4"
	"github.com/upper/db/v4/internal/sqladapter/exql"
	"github.com/upper/db/v4/internal/sqlbuilder"
)

// CollectionAdapter defines methods to be implemented by SQL adapters.
type CollectionAdapter interface {
	// Insert prepares and executes an INSERT statament. When the item is
	// succefully added, Insert returns a unique identifier of the newly added
	// element (or nil if the unique identifier couldn't be determined).
	Insert(Collection, interface{}) (interface{}, error)
}

// Collection satisfies db.Collection.
type Collection interface {
	// Insert inserts a new item into the collection.
	Insert(interface{}) (db.InsertResult, error)

	// Name returns the name of the collection.
	Name() string

	// Session returns the db.Session the collection belongs to.
	Session() db.Session

	// Exists returns true if the collection exists, false otherwise.
	Exists() (bool, error)

	// Find defined a new result set.
	Find(conds ...interface{}) db.Result

	Count() (uint64, error)

	// Truncate removes all elements on the collection and resets the
	// collection's IDs.
	Truncate() error

	// InsertReturning inserts a new item into the collection and refreshes the
	// item with actual data from the database. This is useful to get automatic
	// values, such as timestamps, or IDs.
	InsertReturning(item interface{}) error

	// UpdateReturning updates a record from the collection and refreshes the item
	// with actual data from the database. This is useful to get automatic
	// values, such as timestamps, or IDs.
	UpdateReturning(item interface{}) error

	// PrimaryKeys returns the names of all primary keys in the table.
	PrimaryKeys() ([]string, error)

	// SQLBuilder returns a db.SQL instance.
	SQL() db.SQL
}

type finder interface {
	Find(Collection, *Result, ...interface{}) db.Result
}

type condsFilter interface {
	FilterConds(...interface{}) []interface{}
}

// collection is the implementation of Collection.
type collection struct {
	name    string
	adapter CollectionAdapter
}

type collectionWithSession struct {
	*collection

	session Session
}

func newCollection(name string, adapter CollectionAdapter) *collection {
	if adapter == nil {
		panic("upper: nil adapter")
	}
	return &collection{
		name:    name,
		adapter: adapter,
	}
}

func (c *collectionWithSession) SQL() db.SQL {
	return c.session.SQL()
}

func (c *collectionWithSession) Session() db.Session {
	return c.session
}

func (c *collectionWithSession) Name() string {
	return c.name
}

func (c *collectionWithSession) Count() (uint64, error) {
	return c.Find().Count()
}

func (c *collectionWithSession) Insert(item interface{}) (db.InsertResult, error) {
	id, err := c.adapter.Insert(c, item)
	if err != nil {
		return nil, err
	}

	return db.NewInsertResult(id), nil
}

func (c *collectionWithSession) PrimaryKeys() ([]string, error) {
	return c.session.PrimaryKeys(c.Name())
}

func (c *collectionWithSession) filterConds(conds ...interface{}) ([]interface{}, error) {
	pk, err := c.PrimaryKeys()
	if err != nil {
		return nil, err
	}
	if len(conds) == 1 && len(pk) == 1 {
		if id := conds[0]; IsKeyValue(id) {
			conds[0] = db.Cond{pk[0]: db.Eq(id)}
		}
	}
	if tr, ok := c.adapter.(condsFilter); ok {
		return tr.FilterConds(conds...), nil
	}
	return conds, nil
}

func (c *collectionWithSession) Find(conds ...interface{}) db.Result {
	filteredConds, err := c.filterConds(conds...)
	if err != nil {
		res := &Result{}
		res.setErr(err)
		return res
	}

	res := NewResult(
		c.session.SQL(),
		c.Name(),
		filteredConds,
	)
	if f, ok := c.adapter.(finder); ok {
		return f.Find(c, res, conds...)
	}
	return res
}

func (c *collectionWithSession) Exists() (bool, error) {
	if err := c.session.TableExists(c.Name()); err != nil {
		return false, err
	}
	return true, nil
}

func (c *collectionWithSession) InsertReturning(item interface{}) error {
	if item == nil || reflect.TypeOf(item).Kind() != reflect.Ptr {
		return fmt.Errorf("Expecting a pointer but got %T", item)
	}

	// Grab primary keys
	pks, err := c.PrimaryKeys()
	if err != nil {
		return err
	}

	if len(pks) == 0 {
		if ok, err := c.Exists(); !ok {
			return err
		}
		return fmt.Errorf(db.ErrMissingPrimaryKeys.Error(), c.Name())
	}

	var tx Session
	isTransaction := c.session.IsTransaction()
	if isTransaction {
		tx = c.session
	} else {
		var err error
		tx, err = c.session.NewTransaction(c.session.Context(), nil)
		if err != nil {
			return err
		}
		defer tx.Close()
	}

	// Allocate a clone of item.
	newItem := reflect.New(reflect.ValueOf(item).Elem().Type()).Interface()
	var newItemFieldMap map[string]reflect.Value

	itemValue := reflect.ValueOf(item)

	col := tx.Collection(c.Name())

	// Insert item as is and grab the returning ID.
	var newItemRes db.Result
	id, err := col.Insert(item)
	if err != nil {
		goto cancel
	}
	if id == nil {
		err = fmt.Errorf("InsertReturning: Could not get a valid ID after inserting. Does the %q table have a primary key?", c.Name())
		goto cancel
	}

	if len(pks) > 1 {
		newItemRes = col.Find(id)
	} else {
		// We have one primary key, build a explicit db.Cond with it to prevent
		// string keys to be considered as raw conditions.
		newItemRes = col.Find(db.Cond{pks[0]: id}) // We already checked that pks is not empty, so pks[0] is defined.
	}

	// Fetch the row that was just interted into newItem
	err = newItemRes.One(newItem)
	if err != nil {
		goto cancel
	}

	switch reflect.ValueOf(newItem).Elem().Kind() {
	case reflect.Struct:
		// Get valid fields from newItem to overwrite those that are on item.
		newItemFieldMap = sqlbuilder.Mapper.ValidFieldMap(reflect.ValueOf(newItem))
		for fieldName := range newItemFieldMap {
			sqlbuilder.Mapper.FieldByName(itemValue, fieldName).Set(newItemFieldMap[fieldName])
		}
	case reflect.Map:
		newItemV := reflect.ValueOf(newItem).Elem()
		itemV := reflect.ValueOf(item)
		if itemV.Kind() == reflect.Ptr {
			itemV = itemV.Elem()
		}
		for _, keyV := range newItemV.MapKeys() {
			itemV.SetMapIndex(keyV, newItemV.MapIndex(keyV))
		}
	default:
		err = fmt.Errorf("InsertReturning: expecting a pointer to map or struct, got %T", newItem)
		goto cancel
	}

	if !isTransaction {
		// This is only executed if t.Session() was **not** a transaction and if
		// sess was created with sess.NewTransaction().
		return tx.Commit()
	}

	return err

cancel:
	// This goto label should only be used when we got an error within a
	// transaction and we don't want to continue.

	if !isTransaction {
		// This is only executed if t.Session() was **not** a transaction and if
		// sess was created with sess.NewTransaction().
		_ = tx.Rollback()
	}
	return err
}

func (c *collectionWithSession) UpdateReturning(item interface{}) error {
	if item == nil || reflect.TypeOf(item).Kind() != reflect.Ptr {
		return fmt.Errorf("Expecting a pointer but got %T", item)
	}

	// Grab primary keys
	pks, err := c.PrimaryKeys()
	if err != nil {
		return err
	}

	if len(pks) == 0 {
		if ok, err := c.Exists(); !ok {
			return err
		}
		return fmt.Errorf(db.ErrMissingPrimaryKeys.Error(), c.Name())
	}

	var tx Session
	isTransaction := c.session.IsTransaction()

	if isTransaction {
		tx = c.session
	} else {
		// Not within a transaction, let's create one.
		var err error
		tx, err = c.session.NewTransaction(c.session.Context(), nil)
		if err != nil {
			return err
		}
		defer tx.Close()
	}

	// Allocate a clone of item.
	defaultItem := reflect.New(reflect.ValueOf(item).Elem().Type()).Interface()
	var defaultItemFieldMap map[string]reflect.Value

	itemValue := reflect.ValueOf(item)

	conds := db.Cond{}
	for _, pk := range pks {
		conds[pk] = db.Eq(sqlbuilder.Mapper.FieldByName(itemValue, pk).Interface())
	}

	col := tx.(Session).Collection(c.Name())

	err = col.Find(conds).Update(item)
	if err != nil {
		goto cancel
	}

	if err = col.Find(conds).One(defaultItem); err != nil {
		goto cancel
	}

	switch reflect.ValueOf(defaultItem).Elem().Kind() {
	case reflect.Struct:
		// Get valid fields from defaultItem to overwrite those that are on item.
		defaultItemFieldMap = sqlbuilder.Mapper.ValidFieldMap(reflect.ValueOf(defaultItem))
		for fieldName := range defaultItemFieldMap {
			sqlbuilder.Mapper.FieldByName(itemValue, fieldName).Set(defaultItemFieldMap[fieldName])
		}
	case reflect.Map:
		defaultItemV := reflect.ValueOf(defaultItem).Elem()
		itemV := reflect.ValueOf(item)
		if itemV.Kind() == reflect.Ptr {
			itemV = itemV.Elem()
		}
		for _, keyV := range defaultItemV.MapKeys() {
			itemV.SetMapIndex(keyV, defaultItemV.MapIndex(keyV))
		}
	default:
		panic("default")
	}

	if !isTransaction {
		// This is only executed if t.Session() was **not** a transaction and if
		// sess was created with sess.NewTransaction().
		return tx.Commit()
	}
	return err

cancel:
	// This goto label should only be used when we got an error within a
	// transaction and we don't want to continue.

	if !isTransaction {
		// This is only executed if t.Session() was **not** a transaction and if
		// sess was created with sess.NewTransaction().
		_ = tx.Rollback()
	}
	return err
}

func (c *collectionWithSession) Truncate() error {
	stmt := exql.Statement{
		Type:  exql.Truncate,
		Table: exql.TableWithName(c.Name()),
	}
	if _, err := c.session.SQL().Exec(&stmt); err != nil {
		return err
	}
	return nil
}
