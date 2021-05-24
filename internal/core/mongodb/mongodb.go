package mongodb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	db     *mongo.Database
	client *mongo.Client
	// ErrorNotFound error not found
	ErrorNotFound = errors.New("Not found")

	// ErrorInvalidID error invalid id
	ErrorInvalidID = errors.New("Invalid ID")

	// ErrorDocumentDuplicate error document is duplicate
	ErrorDocumentDuplicate = errors.New("Document is duplicate")

	// ErrorSliceIsEmpty error slice is empty
	ErrorSliceIsEmpty = errors.New("Slice is empty")
)

// Options mongo option
type Options struct {
	URL              string
	Port             int
	DatabaseName     string
	Username         string
	Password         string
	Debug            bool
	HandleNullValues []interface{}
}

var defaultNullValues = []interface{}{
	"",
	int(0),
}

// InitDatabase new database
func InitDatabase(o *Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	uri := fmt.Sprintf("mongodb://%s:%d", o.URL, o.Port)
	if o.Username != "" && o.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?connect=direct", o.Username, o.Password, o.URL, o.Port, o.DatabaseName)
	}
	clientOptions := options.Client().ApplyURI(uri).SetRegistry(buildNullValueDecoder(append(defaultNullValues, o.HandleNullValues)...))
	if o.Debug {
		clientOptions.Monitor = &event.CommandMonitor{
			Started: func(c context.Context, e *event.CommandStartedEvent) {
				fmt.Printf("\033[0;36mMongoDB command exec\033[0m: \033[1;95m%s\033[0m\n", e.Command.String())
			},
		}
	}
	c, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	err = c.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	db = c.Database(o.DatabaseName)
	client = c
	return nil
}

func buildNullValueDecoder(val ...interface{}) *bsoncodec.Registry {
	rb := bson.NewRegistryBuilder()
	for _, v := range val {
		t := reflect.TypeOf(v)
		defDecoder, err := bson.DefaultRegistry.LookupDecoder(t)
		if err != nil {
			panic(err)
		}
		rb.RegisterDecoder(t, &nullValueDecoder{defDecoder, reflect.Zero(t)})
	}
	return rb.Build()
}

type nullValueDecoder struct {
	defDecoder bsoncodec.ValueDecoder
	zeroValue  reflect.Value
}

func (d *nullValueDecoder) DecodeValue(dctx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if vr.Type() != bsontype.Null {
		return d.defDecoder.DecodeValue(dctx, vr, val)
	}
	if !val.CanSet() {
		return errors.New("value not settable")
	}
	if err := vr.ReadNull(); err != nil {
		return err
	}
	val.Set(d.zeroValue)
	return nil
}

// DB database
func DB() *mongo.Database {
	return db
}

// Client database
func Client() *mongo.Client {
	return client
}

// Repo common repo
type Repo struct {
	Collection *mongo.Collection
	Mux        sync.Mutex
}

// Create create user
func (r *Repo) Create(i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if m, ok := i.(ModelInterface); ok {
		if m.GetCreatedAt().IsZero() {
			m.Stamp()
		}
		if m.GetID().IsZero() {
			m.SetID(primitive.NewObjectID())
		}
	}
	_, err := r.Collection.InsertOne(ctx, i)
	if err != nil {
		return wrapError(err)
	}
	return nil
}

// CreateMany create many
func (r *Repo) CreateMany(i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	iV := reflect.ValueOf(i)
	ins := make([]interface{}, 0, iV.Len())
	for j := 0; j < iV.Len(); j++ {
		if m, ok := iV.Index(j).Interface().(ModelInterface); ok {
			if m.GetCreatedAt().IsZero() {
				m.Stamp()
			}
			if m.GetID().IsZero() {
				m.SetID(primitive.NewObjectID())
			}
			ins = append(ins, m)
		}
	}
	rr, err := r.Collection.InsertMany(ctx, ins)
	_ = rr
	if err != nil {
		return wrapError(err)
	}
	return nil
}

// Update update
func (r *Repo) Update(i interface{}) error {
	return r.UpdateByPrimitiveM(primitive.M{
		"$set": i,
	}, i)
}

// UpdateWithoutTimestamp update post without timestamp
func (r *Repo) UpdateWithoutTimestamp(i interface{}) error {
	var id primitive.ObjectID
	if m, ok := i.(ModelInterface); ok {
		id = m.GetID()
	} else {
		return ErrorInvalidID
	}

	s := primitive.M{
		"_id": id,
	}

	u := primitive.M{
		"$set": i,
	}

	if err := r.UpdateOneByPrimitiveM(s, u); err != nil {
		return err
	}

	return nil
}

// Replace replace one
func (r *Repo) Replace(i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var id primitive.ObjectID
	if m, ok := i.(ModelInterface); ok {
		m.UpdateStamp()
		id = m.GetID()
	}
	r.Mux.Lock()
	_, err := r.Collection.ReplaceOne(ctx,
		primitive.D{
			primitive.E{
				Key:   "_id",
				Value: id,
			},
		}, i)
	r.Mux.Unlock()
	if err != nil {
		return err
	}
	return nil
}

// Delete soft delete entity
func (r *Repo) Delete(i interface{}) error {
	if m, ok := i.(ModelInterface); ok {
		m.DeleteStamp()
	}

	return r.Update(i)
}

// HardDelete hard delete entity
func (r *Repo) HardDelete(i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var id primitive.ObjectID
	if m, ok := i.(ModelInterface); ok {
		id = m.GetID()
	}

	d := primitive.M{
		"_id": id,
	}

	r.Mux.Lock()
	_, err := r.Collection.DeleteOne(ctx, d)
	r.Mux.Unlock()
	if err != nil {
		return err
	}
	return nil
}

// HardDeleteAllByPrimitiveM hard delete all by primitive M
func (r *Repo) HardDeleteAllByPrimitiveM(s primitive.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.Mux.Lock()
	_, err := r.Collection.DeleteMany(ctx, s)
	r.Mux.Unlock()

	if err != nil {
		return err
	}

	return nil
}

// Upsert upsert
func (r *Repo) Upsert(i interface{}, s primitive.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var id primitive.ObjectID
	if m, ok := i.(ModelInterface); ok {
		id = m.GetID()
		if !id.IsZero() {
			if s == nil {
				s = primitive.M{}
			}
			s["_id"] = id
			m.UpdateStamp()
		} else {
			m.Stamp()
		}
	}
	r.Mux.Lock()
	_, err := r.Collection.UpdateOne(ctx,
		s, primitive.M{
			"$set": i,
		}, options.Update().SetUpsert(true))
	r.Mux.Unlock()

	if err != nil {
		return err
	}
	return nil
}

// UpsertBySelectorAndUpdate upsert by selector and update
func (r *Repo) UpsertBySelectorAndUpdate(s primitive.M, u primitive.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.Mux.Lock()
	_, err := r.Collection.UpdateOne(ctx, s, u, options.Update().SetUpsert(true))
	r.Mux.Unlock()

	if err != nil {
		return err
	}
	return nil
}

// UnsetFields unset fields
func (r *Repo) UnsetFields(i interface{}, fields []string) error {
	unset := primitive.M{}
	for _, field := range fields {
		unset[field] = 1
	}

	return r.UpdateByPrimitiveM(primitive.M{
		"$unset": unset,
	}, i)
}

// GetReplaceRootWithField get replace root with field
func (r *Repo) GetReplaceRootWithField(field interface{}) primitive.M {
	return primitive.M{
		"$replaceRoot": primitive.M{
			"newRoot": primitive.M{
				"$mergeObjects": primitive.A{
					field,
					"$$ROOT",
				},
			},
		},
	}
}

// AddToSet add to set
func (r *Repo) AddToSet(field string, value interface{}, i interface{}) error {
	return r.toggleInSet("$addToSet", field, value, i)
}

// RemoveFromSet remove from set
func (r *Repo) RemoveFromSet(field string, value interface{}, i interface{}) error {
	return r.toggleInSet("$pull", field, value, i)
}

func (r *Repo) toggleInSet(action string, field string, value interface{}, i interface{}) error {
	q := primitive.M{
		action: primitive.M{
			field: value,
		},
	}

	return r.UpdateByPrimitiveM(q, i)
}

// Inc increment a path with value
func (r *Repo) Inc(id primitive.ObjectID, path string, value int) error {
	err := r.UpdateByPrimitiveM(primitive.M{
		"$inc": primitive.M{
			path: value,
		},
	}, id)
	return err
}

// UpdateByPrimitiveM Update By Primitive M
func (r *Repo) UpdateByPrimitiveM(m primitive.M, i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var id primitive.ObjectID
	if m, ok := i.(ModelInterface); ok {
		m.UpdateStamp()
		id = m.GetID()
	} else if oid, ok := i.(primitive.ObjectID); ok {
		id = oid
	}
	r.Mux.Lock()
	_, err := r.Collection.UpdateOne(ctx,
		primitive.D{
			primitive.E{
				Key:   "_id",
				Value: id,
			},
		}, m)
	r.Mux.Unlock()
	if err != nil {
		return err
	}
	return nil
}

// UpdateManyByPrimitiveM update many by primitive M
func (r *Repo) UpdateManyByPrimitiveM(s primitive.M, u primitive.M) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.Mux.Lock()
	result, err := r.Collection.UpdateMany(ctx, s, u)
	r.Mux.Unlock()

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateOneByPrimitiveM update many by primitive M
func (r *Repo) UpdateOneByPrimitiveM(s primitive.M, u primitive.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	r.Mux.Lock()
	_, err := r.Collection.UpdateOne(ctx, s, u)
	r.Mux.Unlock()

	if err != nil {
		return err
	}

	return nil
}

// FindOneByPrimitiveD find one by primitive.D
func (r *Repo) FindOneByPrimitiveD(d primitive.D, i interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if d == nil {
		d = primitive.D{}
	}
	err := r.Collection.FindOne(ctx, d).Decode(i)
	if err != nil {
		return ErrorNotFound
	}
	return nil
}

// FindOneByPrimitiveM find one by primitive.M
func (r *Repo) FindOneByPrimitiveM(m primitive.M, i interface{}, opts ...*options.FindOneOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := r.Collection.FindOne(ctx, m, opts...).Decode(i)
	if err != nil {
		return ErrorNotFound
	}
	return nil
}

// FindOneByID find one by id
func (r *Repo) FindOneByID(id string, i interface{}) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrorInvalidID
	}
	err = r.FindOneByPrimitiveD(primitive.D{
		primitive.E{
			Key:   "_id",
			Value: oid,
		},
	}, i)
	if err != nil {
		return err
	}
	return nil
}

// FindAll find all
func (r *Repo) FindAll(m primitive.M, result interface{}, opts ...*options.FindOptions) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cur, err := r.Collection.Find(ctx, m, opts...)
	if err != nil {
		return wrapError(err)
	}
	defer func() {
		cerr := cur.Close(ctx)
		if err == nil {
			err = cerr
		}
	}()

	resultv := reflect.ValueOf(result)
	slicev := resultv.Elem()
	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0

	for {
		elemp := reflect.New(elemt)
		if !cur.Next(ctx) {
			break
		}
		err = cur.Decode(elemp.Interface())
		if err != nil {
			return err
		}
		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}

// FindAllByIDs find all by ids
func (r *Repo) FindAllByIDs(ids []string, i interface{}) error {
	oid := r.ConvertStringToPrimitiveObjectIDs(ids)

	selector := primitive.M{
		"_id": primitive.M{
			"$in": oid,
		},
	}

	if err := r.FindAll(selector, i); err != nil {
		return err
	}
	return nil
}

// AggregateAllByPrimitiveA aggregate with pipeline by using primitive A
func (r *Repo) AggregateAllByPrimitiveA(p primitive.A, result interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // TODO: Just for tester to test other issue
	defer cancel()
	opts := options.Aggregate()
	cur, err := r.Collection.Aggregate(ctx, p, opts)
	if err != nil {
		return wrapError(err)
	}
	defer func() {
		cerr := cur.Close(ctx)
		if err == nil {
			err = cerr
		}
	}()

	resultv := reflect.ValueOf(result)
	slicev := resultv.Elem()
	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	i := 0

	for {
		elemp := reflect.New(elemt)
		if !cur.Next(ctx) {
			break
		}

		if err = cur.Decode(elemp.Interface()); err != nil {
			return err
		}

		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}

	if err = cur.Err(); err != nil {
		return err
	}

	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}

// AggregateOneByPrimitiveA aggregate one with pipeline by using primitive A
func (r *Repo) AggregateOneByPrimitiveA(p primitive.A, result interface{}) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	opts := options.Aggregate()
	cur, err := r.Collection.Aggregate(ctx, p, opts)
	if err != nil {
		return wrapError(err)
	}
	defer func() {
		cerr := cur.Close(ctx)
		if err == nil {
			err = cerr
		}
	}()

	resultv := reflect.ValueOf(result)
	elemt := resultv.Type().Elem()

	elemp := reflect.New(elemt)
	if !cur.Next(ctx) {
		return ErrorNotFound
	}
	if err = cur.Decode(elemp.Interface()); err != nil {
		return err
	}

	if err = cur.Err(); err != nil {
		return err
	}

	resultv.Elem().Set(elemp.Elem())
	return nil
}

// CountDocumentByPrimitiveM count document by primitive.M
func (r *Repo) CountDocumentByPrimitiveM(m primitive.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	count, err := r.Collection.CountDocuments(ctx, m)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetSort get sort
func (r *Repo) GetSort(field string, direction int) primitive.M {
	return primitive.M{
		"$sort": primitive.M{
			field: direction,
		},
	}
}

// GetRegex build Regx
func GetRegex(field string, keyword string, options string) primitive.M {
	return primitive.M{
		field: primitive.M{
			"$regex":   keyword,
			"$options": options,
		},
	}
}

// GetRange get range
func GetRange(field string, from interface{}, to interface{}) primitive.M {
	return primitive.M{
		field: primitive.M{
			"$gte": from,
			"$lte": to,
		},
	}
}

// GetLookup Get Look Up
func (r *Repo) GetLookup(collection string, localField string, foreignID string, as string) primitive.M {
	return primitive.M{
		"$lookup": primitive.M{
			"from":         collection,
			"localField":   localField,
			"foreignField": foreignID,
			"as":           as,
		},
	}
}

// GetUnwind Get unwind
func (r *Repo) GetUnwind(path string, preserve bool) primitive.M {
	return primitive.M{
		"$unwind": primitive.M{
			"path":                       path,
			"preserveNullAndEmptyArrays": preserve,
		},
	}
}

// GetGroup get $group
func (r *Repo) GetGroup(id string, fields map[string]primitive.M) primitive.M {
	group := primitive.M{
		"_id": id,
	}

	for n, v := range fields {
		group[n] = v
	}

	return primitive.M{
		"$group": group,
	}
}

// GetProject get project
func (r *Repo) GetProject(fields map[string]interface{}) primitive.M {
	project := primitive.M{}

	for n, v := range fields {
		project[n] = v
	}

	return primitive.M{
		"$project": project,
	}
}

// GetIDsFromItems get ids for items
func (r *Repo) GetIDsFromItems(itemValues interface{}, idField string) []primitive.ObjectID {
	items := reflect.ValueOf(itemValues)
	var ids []primitive.ObjectID
	ids = []primitive.ObjectID{}
	seen := map[primitive.ObjectID]bool{}
	for index := 0; index < items.Len(); index++ {
		idValue := reflect.
			Indirect(items.Index(index)).
			FieldByName(idField)

		id := idValue.Interface()
		var oid primitive.ObjectID
		if idValue.Kind() == reflect.Ptr {
			oid = *(id.(*primitive.ObjectID))
		} else {
			oid = id.(primitive.ObjectID)
		}
		if idValue.IsValid() && id != nil && !oid.IsZero() {
			if _, ok := seen[oid]; !ok {
				seen[oid] = true
				ids = append(ids, oid)
			}
		}
	}
	return ids
}

// ConvertStringToPrimitiveObjectIDs convert []string to  []primitive.ObjectID
func (r *Repo) ConvertStringToPrimitiveObjectIDs(ids []string) []primitive.ObjectID {
	var objectIds []primitive.ObjectID
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objectIds = append(objectIds, oid)
		}
	}
	return objectIds
}

// SubtractIDSlices diff idslices
func SubtractIDSlices(a, b []primitive.ObjectID) []primitive.ObjectID {
	r := []primitive.ObjectID{}

	for _, i := range a {
		add := true
		for _, j := range b {
			if i == j {
				add = false
				break
			}
		}
		if add {
			r = append(r, i)
		}
	}

	return r
}

// SliceToIDMapGroup slice to idmap group
func SliceToIDMapGroup(slice interface{}, idField string, limit int) map[primitive.ObjectID]interface{} {
	sliceV := reflect.ValueOf(slice)
	m := map[primitive.ObjectID]interface{}{}
	groupsLimitMap := map[primitive.ObjectID]int{}

	if sliceV.Len() == 0 || limit < 1 {
		return m
	}
	valueMap := map[primitive.ObjectID][]reflect.Value{}

	for i := 0; i < sliceV.Len(); i++ {
		idV := reflect.
			Indirect(sliceV.Index(i)).
			FieldByName(idField)

		id := idV.Interface()
		var oid primitive.ObjectID
		if idV.Kind() == reflect.Ptr {
			oid = *(id.(*primitive.ObjectID))
		} else {
			oid = id.(primitive.ObjectID)
		}
		if idV.IsValid() && id != nil && !oid.IsZero() {
			if limit > 0 && groupsLimitMap[oid] >= limit {
				continue
			}
			valueMap[oid] = append(valueMap[oid], sliceV.Index(i))
			groupsLimitMap[oid]++
		}
	}

	for k, vSlice := range valueMap {
		m[k] = reflect.MakeSlice(reflect.SliceOf(vSlice[0].Type()), len(vSlice), len(vSlice))
		for i, v := range vSlice {
			m[k].(reflect.Value).Index(i).Set(v)
		}
		m[k] = m[k].(reflect.Value).Interface()
	}

	return m
}

// SliceToIDMap object slice to id map
func SliceToIDMap(slice interface{}, idField string) map[primitive.ObjectID]interface{} {
	sliceV := reflect.ValueOf(slice)
	m := make(map[primitive.ObjectID]interface{}, sliceV.Len())
	if sliceV.Len() == 0 {
		return m
	}

	for i := 0; i < sliceV.Len(); i++ {
		idV := reflect.
			Indirect(sliceV.Index(i)).
			FieldByName(idField)
		id := idV.Interface()
		var oid primitive.ObjectID
		if idV.Kind() == reflect.Ptr {
			oid = *(id.(*primitive.ObjectID))
		} else {
			oid = id.(primitive.ObjectID)
		}
		if idV.IsValid() && id != nil && !oid.IsZero() {
			m[oid] = sliceV.Index(i).Interface()
		}
	}

	return m
}

// wrapError wrap error
func wrapError(err error) error {
	if e, ok := err.(mongo.WriteException); ok {
		if len(e.WriteErrors) > 0 {
			we := e.WriteErrors[0]
			switch we.Code {
			case 11000:
				return ErrorDocumentDuplicate
			default:
				return fmt.Errorf("(%d) %s", we.Code, we.Message)
			}
		}
	}
	return nil
}
