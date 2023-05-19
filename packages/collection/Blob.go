package collection

import "gorm.io/gorm"

type Blob struct {
	gorm.Model
	Data []byte `gorm:"type:blob"`
	dbMethods
}

func blobNew(data []byte) (*Blob, error) {
	// Create new object, save to db, return it
	o := &Blob{
		Data: data,
	}
	err := o.save(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func blobLoad(blobID uint) (*Blob, error) {
	o := &Blob{}
	// The following db.Where... will not do nested-preloading, that will be done latter
	// with the o.reload(o) call
	err := db.First(o, "id = ?", blobID).Error
	if err != nil {
		return nil, err
	}
	err = o.reload(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func blobData(blobID uint) (data []byte, err error) {
	blob, err := blobLoad(blobID)
	return blob.Data, err

}

func (o *Blob) gormID() uint {
	return o.ID
}
