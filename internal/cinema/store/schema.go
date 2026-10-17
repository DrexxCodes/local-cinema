package store

import (
	"github.com/pocketbase/pocketbase/core"
)

// autodates adds the created and updated timestamps every collection gets.
func autodates() []core.Field {
	return []core.Field{
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	}
}

// EnsureCollections creates the collections and fields the program needs, and adds fields
// that databases made by older versions are missing. It is safe to run on every start.
func (s *Store) EnsureCollections() error {
	users, err := s.App.FindCollectionByNameOrId("users")
	if err != nil {
		return err
	}
	dirty := false
	if users.Fields.GetByName("room_number") == nil {
		users.Fields.Add(
			&core.TextField{Name: "room_number"},
			&core.TextField{Name: "display_name"},
			&core.NumberField{Name: "max_connections"},
			&core.BoolField{Name: "active"},
			&core.TextField{Name: "notes"},
		)
		dirty = true
	}
	// Guests never talk to PocketBase's REST API directly. Everything goes
	// through our own routes, so lock the collection (superusers bypass rules).
	if users.ListRule != nil || users.ViewRule != nil || users.CreateRule != nil ||
		users.UpdateRule != nil || users.DeleteRule != nil {
		users.ListRule, users.ViewRule, users.CreateRule = nil, nil, nil
		users.UpdateRule, users.DeleteRule = nil, nil
		dirty = true
	}
	if dirty {
		if err := s.App.Save(users); err != nil {
			return err
		}
	}

	if err := s.ensure("settings", func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "cinema_name"},
			&core.TextField{Name: "tagline"},
			&core.TextField{Name: "accent_color"},
			&core.JSONField{Name: "movie_folders", MaxSize: 1 << 20},
			&core.BoolField{Name: "watch_folders"},
			&core.NumberField{Name: "global_max_streams"},
			&core.TextField{Name: "mdns_name"},
			&core.BoolField{Name: "setup_complete"},
			&core.NumberField{Name: "setup_step"},
			&core.TextField{Name: "tmdb_api_key"},
		)
	}); err != nil {
		return err
	}

	if err := s.ensure("movies", func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "title"},
			&core.NumberField{Name: "year"},
			&core.TextField{Name: "file_path"},
			&core.NumberField{Name: "file_size"},
			&core.NumberField{Name: "duration_seconds"},
			&core.TextField{Name: "container"},
			&core.TextField{Name: "video_codec"},
			&core.TextField{Name: "audio_codec"},
			&core.TextField{Name: "resolution"},
			&core.TextField{Name: "synopsis"},
			&core.JSONField{Name: "genres", MaxSize: 1 << 16},
			&core.NumberField{Name: "rating"},
			&core.NumberField{Name: "tmdb_id"},
			&core.BoolField{Name: "has_poster"},
			&core.BoolField{Name: "has_backdrop"},
			&core.BoolField{Name: "available"},
			&core.DateField{Name: "added_at"},
			&core.DateField{Name: "last_scanned"},
		)
		c.AddIndex("idx_movies_path", true, "file_path", "")
	}); err != nil {
		return err
	}

	if err := s.ensure("sessions", func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, CascadeDelete: true},
			&core.TextField{Name: "device_id"},
			&core.TextField{Name: "device_name"},
			&core.TextField{Name: "ip"},
			&core.TextField{Name: "user_agent"},
			&core.TextField{Name: "room_number"},
			&core.DateField{Name: "started_at"},
			&core.DateField{Name: "last_seen"},
			&core.BoolField{Name: "active"},
		)
		c.AddIndex("idx_sessions_user_active", false, "user,active", "")
	}); err != nil {
		return err
	}

	// Schema upgrades for databases created by older versions.
	if err := s.addMissingFields("users", &core.BoolField{Name: "allow_external"}); err != nil {
		return err
	}
	if err := s.addMissingFields("sessions",
		&core.NumberField{Name: "rtt_ms"},
		&core.NumberField{Name: "signal"},
		&core.TextField{Name: "kind"}, // "" for a browser, "external" for a stream URL used in another app
	); err != nil {
		return err
	}
	if err := s.addMissingFields("movies",
		&core.NumberField{Name: "thumb_version"},
		&core.BoolField{Name: "hidden"}, // admin removed it from the stream
		&core.BoolField{Name: "edited"}, // admin typed the details, so scans must not overwrite them
	); err != nil {
		return err
	}

	movies, err := s.App.FindCollectionByNameOrId("movies")
	if err != nil {
		return err
	}
	if err := s.ensure("playback_state", func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, CascadeDelete: true},
			&core.RelationField{Name: "movie", CollectionId: movies.Id, MaxSelect: 1, CascadeDelete: true},
			&core.NumberField{Name: "position_seconds"},
			&core.BoolField{Name: "completed"},
		)
		c.AddIndex("idx_playback_user_movie", true, "user,movie", "")
	}); err != nil {
		return err
	}

	if err := s.ensure("upgrade_requests", func(c *core.Collection) {
		c.Fields.Add(
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, CascadeDelete: true},
			&core.NumberField{Name: "slots"},   // extra connections the guest asked for
			&core.NumberField{Name: "granted"}, // what the admin approved
			&core.TextField{Name: "status"},    // pending, approved or declined
		)
	}); err != nil {
		return err
	}

	return s.ensure("audit_log", func(c *core.Collection) {
		c.Fields.Add(
			&core.TextField{Name: "event"},
			&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1},
			&core.TextField{Name: "session"},
			&core.TextField{Name: "movie"},
			&core.JSONField{Name: "detail", MaxSize: 1 << 16},
		)
	})
}

// ensure creates a collection when it does not exist yet.
func (s *Store) ensure(name string, build func(*core.Collection)) error {
	if _, err := s.App.FindCollectionByNameOrId(name); err == nil {
		return nil
	}
	c := core.NewBaseCollection(name)
	build(c)
	c.Fields.Add(autodates()...)
	return s.App.Save(c)
}

// addMissingFields adds fields to an existing collection when they are absent.
func (s *Store) addMissingFields(name string, fields ...core.Field) error {
	c, err := s.App.FindCollectionByNameOrId(name)
	if err != nil {
		return err
	}
	dirty := false
	for _, f := range fields {
		if c.Fields.GetByName(f.GetName()) == nil {
			c.Fields.Add(f)
			dirty = true
		}
	}
	if !dirty {
		return nil
	}
	return s.App.Save(c)
}
