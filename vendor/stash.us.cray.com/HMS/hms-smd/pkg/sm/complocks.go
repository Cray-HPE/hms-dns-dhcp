// Copyright (c) 2019 Cray Inc. All Rights Reserved.
//
// Except as permitted by contract or express written permission of Cray Inc.,
// no part of this work or its content may be modified, used, reproduced or
// disclosed in any form. Modifications made without express permission of
// Cray Inc. may damage the system the software is installed within, may
// disqualify the user from receiving support from Cray Inc. under support or
// maintenance contracts, or require additional support services outside the
// scope of those contracts to repair the software or system.
//
// This file is contains struct defines for CompLocks

package sm

// This package defines structures for component locks

import (
	base "stash.us.cray.com/HMS/hms-base"
)

//
// Format checking for database keys and query parameters.
//

var ErrCompLockBadLifetime = base.NewHMSError("sm",
	"Invalid CompLock lifetime")

///////////////////////////////////////////////////////////////////////////
//
// CompLock
//
///////////////////////////////////////////////////////////////////////////

// A component lock is a formal, non-overlapping group of components that are
// reserved by a service.
type CompLock struct {
	ID       string   `json:"id"`
	Created  string   `json:"created,omitempty"`
	Reason   string   `json:"reason"`
	Owner    string   `json:"owner"`
	Lifetime int      `json:"lifetime"`
	Xnames   []string `json:"xnames"` // List of xname ids, required.

	// Private
	normalized bool
	verified   bool
}

// Allocate and initialize new CompLock struct, validating it.
// If you already have a created CompLock, you can check the inputs with
// CompLock.Verify()
func NewCompLock(reason, owner string, lifetime int, xnames []string) (*CompLock, error) {
	cl := new(CompLock)
	cl.Reason = reason
	cl.Owner = owner
	cl.Lifetime = lifetime
	cl.Xnames = append([]string(nil), xnames...)
	cl.Normalize()
	return cl, cl.Verify()
}

// Normalize xnames in CompLockMembers.
func (cl *CompLock) Normalize() {
	if cl.normalized == true {
		return
	}
	cl.normalized = true

	for i, xname := range cl.Xnames {
		cl.Xnames[i] = base.NormalizeHMSCompID(xname)
	}
}

// Check input fields of a group.  If no error is returned, the result should
// be ok to put into the database.
func (cl *CompLock) Verify() error {
	if cl.verified == true {
		return nil
	}
	cl.verified = true

	if cl.Lifetime <= 0 {
		return ErrCompLockBadLifetime
	}

	for _, xname := range cl.Xnames {
		if ok := base.IsHMSCompIDValid(xname); ok == false {
			return base.ErrHMSTypeInvalid
		}
	}
	return nil
}

// Patchable fields if included in payload.
type CompLockPatch struct {
	Reason   *string
	Owner    *string
	Lifetime *int
}

// Normalize CompLockPatch (just lower case tags, basically, but keeping same
// interface as others.
func (clp *CompLockPatch) Normalize() {
	// Nothing to do
	return
}

// Analgous Verify call for CompLockPatch objects.
func (clp *CompLockPatch) Verify() error {
	if clp.Lifetime != nil && *clp.Lifetime <= 0 {
		return ErrCompLockBadLifetime
	}
	return nil
}
