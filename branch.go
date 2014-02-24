package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
#include <git2/errors.h>
*/
import "C"

import (
	"unsafe"
)

type BranchType uint

const (
	BranchLocal  BranchType = C.GIT_BRANCH_LOCAL
	BranchRemote            = C.GIT_BRANCH_REMOTE
)

const (
	RefsDir        = "refs/"
	RefsHeadsDir   = RefsDir + "heads/"
	RefsTagsDir    = RefsDir + "tags/"
	RefsRemotesDir = RefsDir + "remotes/"
)

type Branch struct {
	Reference
}

func (repo *Repository) CreateBranch(branchName string, target *Commit, force bool, signature *Signature, message string) (*Reference, error) {
	ref := new(Reference)
	cBranchName := C.CString(branchName)
	cForce := cbool(force)

	cSignature := signature.toC()
	defer C.git_signature_free(cSignature)

	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))

	err := C.git_branch_create(&ref.ptr, repo.ptr, cBranchName, target.ptr, cForce, cSignature, cMessage)
	if err < 0 {
		return nil, LastError()
	}
	return ref, nil
}

func (b *Branch) BranchDelete() error {
	if err := C.git_branch_delete(b.ptr); err < 0 {
		return LastError()
	}
	return nil
}

func (b *Branch) Move(newBranchName string, force bool, signature *Signature, message string) (*Branch, error) {
	newBranch := new(Branch)
	cNewBranchName := C.CString(newBranchName)
	cForce := cbool(force)

	cSignature := signature.toC()
	defer C.git_signature_free(cSignature)

	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))

	err := C.git_branch_move(&newBranch.ptr, b.ptr, cNewBranchName, cForce, cSignature, cMessage)
	if err < 0 {
		return nil, LastError()
	}
	return newBranch, nil
}

func (b *Branch) IsHead() (bool, error) {
	isHead := C.git_branch_is_head(b.ptr)
	switch isHead {
	case 1:
		return true, nil
	case 0:
		return false, nil
	default:
		return false, LastError()
	}

}

func (repo *Repository) LookupBranch(branchName string, bt BranchType) (*Branch, error) {
	branch := new(Branch)
	cName := C.CString(branchName)

	err := C.git_branch_lookup(&branch.ptr, repo.ptr, cName, C.git_branch_t(bt))
	if err < 0 {
		return nil, LastError()
	}
	return branch, nil
}

func (b *Branch) Name() (string, error) {
	var cName *C.char
	defer C.free(unsafe.Pointer(cName))

	err := C.git_branch_name(&cName, b.ptr)
	if err < 0 {
		return "", LastError()
	}

	return C.GoString(cName), nil
}

func (repo *Repository) RemoteName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	nameBuf := C.git_buf{}

	if C.git_branch_remote_name(&nameBuf, repo.ptr, cName) < 0 {
		return "", LastError()
	}
	C.git_buf_free(&nameBuf)

	return C.GoStringN(nameBuf.ptr, C.int(nameBuf.size)), nil
}

func (b *Branch) SetUpstream(upstreamName string) error {
	cName := C.CString(upstreamName)

	err := C.git_branch_set_upstream(b.ptr, cName)
	if err < 0 {
		return LastError()
	}
	return nil
}

func (b *Branch) Upstream() (*Branch, error) {
	upstream := new(Branch)
	err := C.git_branch_upstream(&upstream.ptr, b.ptr)
	if err < 0 {
		return nil, LastError()
	}
	return upstream, nil
}

func (repo *Repository) UpstreamName(canonicalBranchName string) (string, error) {
	cName := C.CString(canonicalBranchName)

	nameBuf := C.git_buf{}

	if C.git_branch_upstream_name(&nameBuf, repo.ptr, cName) < 0 {
		return "", LastError()
	}
	C.git_buf_free(&nameBuf)

	return C.GoStringN(nameBuf.ptr, C.int(nameBuf.size)), nil
}
