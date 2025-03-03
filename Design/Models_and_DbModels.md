
## How to model objects with supporting db-structs

This was "T005) MyObj struct -> MyObj struct&Methods +  DbMyObjIfc + DbMyObj db-struct&db-methods"

### Short summary

```
*package*       *type*            *func*
models          Obj              .GetZz()   .SetZz()    (uses DbObjIfc)
dbModels        DbObjIfc         .GetZz()   .SetZz()    (ifc definition)   
                DbObj            .GetZz()   .SetZz()    (ifc implementation, saves/reloads/etc with db)[1]
                DbObj            .Zz                    (db-struct field)

[1] if Zz is a changeable field, then .GetZz()/.SetZz() funcs should internally start with .Reload()'ing from the db, to assure not using outdated values (ie, if another thread has changed the value in the db)
```

MyModel is what the rest of the app needs to use
DbMyModel is an internal thing of MyModel, to encapsulates the db complexities. 

Func names for MyModel and DbMyModel:
- GetZz() / SetZz() / AppendZz()
- IsYYYY() bool
- DoMyAction()




### MyObj
                - focuses on thehigher logic of MyObj, that is expected for other packages to use
                - other packages should use MyObj (and not DbMyObj)
                - Public methods (ex: MyObj.GetZz() MyObj.SetZz()) should internally use DbMyObjIfc.GetZz() DbMyObjIfc.SetZz()), to access r/w data (and never access DbMyObj). 
                - MyObj constructor will create a DbMyObj, which will be stored internally in the MyObj struct, but all MyObj methods should only use DbMyObjIfc and never again DbMyObj direcly 
                - package named "models"
                - aModelA.GetModelBID() string or aModelA.GetModelB() *ModelB
                  if aModelA relates to another ModelB, its ok for ModelA to return ModelBID string or *ModelB.
                  On the other hand, at the dbModel level it should return dbModelBIDuint or *dbModelB
                  The idea is that at model level only ModelX are returned, and at the dbModel level only dbModelX are returned
                - ID: in package "models", each MyObj has an IDstring = <MyObjPrefix> + <dbMyObj-IDuint>
                  Helper methods convert between IDstring (of MyObj) <-> IDuint (of DbMyObj)
                    ConvertMyObjID_2_IDuint(ID string) (IDuint uint, err error)
                    ConvertMyObjIDuint_2_ID(IDuint uint) (ID string)  {... set and use "<prefix>-1234" ...}
                  The DbMyObj.ID uint == ConvertMyObjID_2_IDuint(MyObj.ID) can be used relate MyObj with DbMyObj
                  Other models like ModelB that need to Load a MyObj, can get the MyObjIDuint from  ConvertMyObjID_2_IDuint() and relate to the underlying dbMyObj representing it
                - Minimal methods:
                    MyObjNew(...) *MyObj
                    MyObjLoad(id string) (*MyObj, error)
                    MyObj_convert_ID_2_IDuint(ID string) (IDuint uint, err error)
                    MyObj_convert_IDuint_2_ID(IDuint uint) (ID string)  {... set and use "<prefix>-1234" ...}
                    (*MyObj).GetID() string
                - Field/Ops methods:
                    (*MyObj).GetZz() (Zz, error)
                    (*MyObj).SetZz(newZz) error
                - Optional methods if relates with another ModelB
                    (*MyObj).GetModelB() *ModelB   // if DbMyObj has field *DbModelB
                    or
                    (*MyObj).GetModelBID() string  // if DbMyObj has a field DbModelBID uint

### DbMyObjIfc
                - interface for DbMyObj, exposing method(s) GetZz() SetZz() for fields of the db-struct, and operations on them
                - package named "dbModels"
                - Minimal ifc methods:  
                    GetID() uint
                - Field/Ops methods:
                    GetZz() (Zz, error)
                    SetZz(newZz) error
                - Optional methods if relates with another dbModelB:
                    (*DbMyObj).GetDbModelB() *DbModelB  // if DbMyObj has field *DbModelB
                    or
                    (*DbMyObj).GetDbModelBID() uint     // if DbMyObj has a field DbModelBID uint

### DbMyObj
                - db-struct and db-methods, intended for Gorm, and to deal with db details and operations
                - Must implement DbMyObjIfc that will be used by MyObj. Other packages should never use DbMyObj directly, only MyObj
                - the struct should have exposed fields, as required by Gorm, but should not be used by other packages
                - package named "dbModels"
                - Minimal DbMyObj methods (implements DbMyObjIfc and some other ones):
                    DbMyObjNew(...) *DbMyObj
                    DbMyObjLoad(id uint) (*DbMyObj, error)
                    GetID() uint
                - Field/Ops methods:
                    GetZz() (Zz, error)
                    SetZz(newZz) error
                - Optional methods if relates with another dbModelB:
                    (*DbMyObj).GetDbModelB() *DbModelB  // if DbMyObj has field *DbModelB
                    or
                    (*DbMyObj).GetDbModelBID() uint     // if DbMyObj has a field DbModelBID uint

### MyObj + DbMyObjIfc + DbMyObj
                - new db-fields should be added to db-struct + DbMyObjIfc + MyObj, with GetZzz() and SetZzz() methods
                - MyObj and DbMyObj can be related by the ID: MyObj.MyObj_convert_ID_2_IDuint(MyObj.GetID()) == DbMyObj.GetID(). 
                  This relation matches MyObj instance with DbMyObj instance 

### Example
```go
// Ex:
// Collection struct&Methods + DbCollectionIfc + DbCollection db-struct&db-methods

// Other packages: should use the constructor and public-methods
package anyOtherPackage
col := CollectionNew(...)
... col.Zzzz() ...

----------------------------

// package Collection
import dbModels
package models
type Collection struct {
  dbCollectionIfc DbCollectionIfc       // unexported field, only used by Collection package and not other packages
}
// Constructor creates dbCollectionIfc and public-methods use dbCollectionIfc to access r/w data
func CollectionNew(...) {...create DbCollection into dbCollectionIfc and return Collection...}
func CollectionLoad(colID string) (*Collection, error)
func Collection_convert_ID_2_IDuint(ID string) (IDuint uint, err error)
func Collection_convert_IDuint_2_ID(IDuint uint) (ID string)  {... set and use "<prefix>-1234" ...}

func (c *Collection) GetID() string {...use dbCollectionIfc.GetID()...}
func (c *Collection) GetName() string {...use dbCollectionIfc.GetName()...}
func (c *Collection) Rename(newName string) string {...use dbCollectionIfc.SetName()...}

-----------------------------

package dbModels
// DbCollection follows gorm conventions, takes care of db-struct and db-methods
interface DbCollectionIfc {
	GetID() uint
	GetName() (string, error)
	SetName(string) error
  ...
}
type DbCollection struct {
	gormCrud.GormCrud[DbCollection]
	Name           string
  ...
}

func DbCollectionNew(...) (*DbCollection, error) {...create and return DbCollection...}
func DbCollectionLoad(dbColId uint) (*DbCollection, error) 

func (d *DbCollection) GetID() uint {
  // d.Reload()  // if this field could change in the db, then reload before delivering it. If field is static, no reload needed
  return d.ID
}
func (d *DbCollection) GetName() (string, error){
  d.Reload()     // if this field could change in the db, then reload before delivering it. If field is static, no reload needed
  return d.Name, nil
}
func (d *DbCollection) SetName(newName string) error { d.Reload(), d.Name = newName, d.Save(d) }
``` 

See also a complete reference implementation in files packages/models/collection.go and packages/dbModels/dbCollection.go 

