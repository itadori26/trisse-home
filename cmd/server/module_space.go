package main

import (
	"trisse-home/internal/modules"
	"trisse-home/internal/modules/space"
)

// The space module registers itself; no shared file lists it (#193).
func init() {
	registerModule("space", func(d moduleDeps) modules.Module {
		return space.New(space.Config{DB: d.db, Cipher: d.out.secrets, Client: d.out.client, Clock: d.clock,
			LaunchURL: d.cfg.SpaceLaunchAPIURL, PictureURL: d.cfg.SpacePictureAPIURL})
	})
}
