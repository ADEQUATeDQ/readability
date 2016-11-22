package portalwatch

type CKANMDAustriaRessources struct {
	Created       *string `json:"created"` // actually a time.Time
	Description   *string `json:"description"`
	Format        *string `json:"format"`
	Hash          *string `json:"hash"`
	ID            *string `json:"id"`
	LastModified  *string `json:"last_modified"` // actually a time.Time
	Mimetype      *string `json:"mimetype"`      // actually a time.Time
	MimetypeInner *string `json:"mimetype_inner"`
	Name          *string `json:"name"`
	Size          *int    `json:"size"`
	URL           *string `json:"url"`
}

type CKANMDAustria struct {
	Author        *string `json:"author"`
	AuthorEmail   *string `json:"author_email"`
	CreatorUserID *string `json:"creator_user_id"`
	Extras        []struct {
		Key   *string     `json:"key"`
		Value interface{} `json:"value"`
	} `json:"extras"`
	ID                *string                   `json:"id"`
	Isopen            bool                      `json:"isopen"`
	LicenseID         *string                   `json:"license_id"`
	LicenseTitle      *string                   `json:"license_title"`
	LicenseURL        *string                   `json:"license_url"`
	Maintainer        *string                   `json:"maintainer"`
	MaintainerEmail   *string                   `json:"maintainer_email"`
	MetadataCreated   *string                   `json:"metadata_created"`  // actually a time.Time
	MetadataModified  *string                   `json:"metadata_modified"` // actually a time.Time
	Name              *string                   `json:"name"`
	Notes             *string                   `json:"notes"`
	Resources         []CKANMDAustriaRessources `json:"resources"`
	RevisionID        *string                   `json:"revision_id"`
	RevisionTimestamp *string                   `json:"revision_timestamp"`
	State             *string                   `json:"state"`
	Tags              []struct {
		DisplayName       *string `json:"display_name"`
		ID                *string `json:"id"`
		Name              *string `json:"name"`
		RevisionTimestamp *string `json:"revision_timestamp"`
		State             *string `json:"state"`
		VocabularyID      *string `json:"vocabulary_id"`
	} `json:"tags"`
	Title   *string `json:"title"`
	Type    *string `json:"type"`
	URL     *string `json:"url"`
	Version *string `json:"version"`
}

type PWMetaData struct {
	Created      *string        `json:"created"`
	License      *string        `json:"license"`
	Md5          *string        `json:"md5"`
	Modified     *string        `json:"modified"`
	Organisation *string        `json:"organisation"`
	Raw          *CKANMDAustria `json:"raw"`
}
