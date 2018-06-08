// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// +build linux darwin freebsd

package fuse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/log"
)

var (
	errEmptyMountPoint      = errors.New("need non-empty mount point")
	errNoRelativeMountPoint = errors.New("invalid path for mount point (need absolute path)")
	errMaxMountCount        = errors.New("max FUSE mount count reached")
	errMountTimeout         = errors.New("mount timeout")
	errAlreadyMounted       = errors.New("mount point is already serving")
)

func isFUSEUnsupportedError(err error) bool {
	if perr, ok := err.(*os.PathError); ok {
		return perr.Op == "open" && perr.Path == "/dev/fuse"
	}
	return err == fuse.ErrOSXFUSENotFound
}

// MountInfo contains information about every active mount
type MountInfo struct {
	MountPoint     string
	StartManifest  string
	LatestManifest string
	rootDir        *SwarmDir
	fuseConnection *fuse.Conn
	swarmApi       *api.Api
	lock           *sync.RWMutex
}

func NewMountInfo(mhash, mpoint string, sapi *api.Api) *MountInfo {
	log.Debug("swarmfs NewMountInfo", "hash", mhash, "mount point", mpoint)
	newMountInfo := &MountInfo{
		MountPoint:     mpoint,
		StartManifest:  mhash,
		LatestManifest: mhash,
		rootDir:        nil,
		fuseConnection: nil,
		swarmApi:       sapi,
		lock:           &sync.RWMutex{},
	}
	return newMountInfo
}

func (swarmfs *SwarmFS) Mount(mhash, mountpoint string) (*MountInfo, error) {
	log.Info("swarmfs", "mounting hash", mhash, "mount point", mountpoint)
	if mountpoint == "" {
		return nil, errEmptyMountPoint
	}
	if !strings.HasPrefix(mountpoint, "/") {
		return nil, errNoRelativeMountPoint
	}
	cleanedMountPoint, err := filepath.Abs(filepath.Clean(mountpoint))
	if err != nil {
		return nil, err
	}
	log.Trace("swarmfs mount", "cleanedMountPoint", cleanedMountPoint)

	swarmfs.swarmFsLock.Lock()
	defer swarmfs.swarmFsLock.Unlock()

	noOfActiveMounts := len(swarmfs.activeMounts)
	log.Debug("swarmfs mount", "# active mounts", noOfActiveMounts)
	if noOfActiveMounts >= maxFuseMounts {
		return nil, errMaxMountCount
	}

	if _, ok := swarmfs.activeMounts[cleanedMountPoint]; ok {
		return nil, errAlreadyMounted
	}

	log.Trace("swarmfs mount: getting manifest tree")
	_, manifestEntryMap, err := swarmfs.swarmApi.BuildDirectoryTree(mhash, true)
	if err != nil {
		return nil, err
	}

	log.Trace("swarmfs mount: building mount info")
	mi := NewMountInfo(mhash, cleanedMountPoint, swarmfs.swarmApi)

	dirTree := map[string]*SwarmDir{}
	rootDir := NewSwarmDir("/", mi)
	log.Trace("swarmfs mount", "rootDir", rootDir)
	mi.rootDir = rootDir

	log.Trace("swarmfs mount: traversing manifest map")
	for suffix, entry := range manifestEntryMap {
		addr := common.Hex2Bytes(entry.Hash)
		fullpath := "/" + suffix
		basepath := filepath.Dir(fullpath)
		parentDir := rootDir
		dirUntilNow := ""
		paths := strings.Split(basepath, "/")
		for i := range paths {
			if paths[i] != "" {
				thisDir := paths[i]
				dirUntilNow = dirUntilNow + "/" + thisDir

				if _, ok := dirTree[dirUntilNow]; !ok {
					dirTree[dirUntilNow] = NewSwarmDir(dirUntilNow, mi)
					parentDir.directories = append(parentDir.directories, dirTree[dirUntilNow])
					parentDir = dirTree[dirUntilNow]

				} else {
					parentDir = dirTree[dirUntilNow]
				}
			}
		}
		thisFile := NewSwarmFile(basepath, filepath.Base(fullpath), mi)
		thisFile.addr = addr

		parentDir.files = append(parentDir.files, thisFile)
	}

	fconn, err := fuse.Mount(cleanedMountPoint, fuse.FSName("swarmfs"), fuse.VolumeName(mhash))
	if isFUSEUnsupportedError(err) {
		log.Error("swarmfs error - FUSE not installed", "mountpoint", cleanedMountPoint, "err", err)
		return nil, err
	} else if err != nil {
		fuse.Unmount(cleanedMountPoint)
		log.Error("swarmfs error mounting swarm manifest", "mountpoint", cleanedMountPoint, "err", err)
		return nil, err
	}
	mi.fuseConnection = fconn

	serverr := make(chan error, 1)
	go func() {
		log.Info("swarmfs", "serving hash", mhash, "at", cleanedMountPoint)
		filesys := &SwarmRoot{root: rootDir}
		if err := fs.Serve(fconn, filesys); err != nil {
			log.Warn("swarmfs could not serve the requested hash", "error", err)
			serverr <- err
		}

	}()

	// Check if the mount process has an error to report.
	select {
	case <-time.After(mountTimeout):
		fuse.Unmount(cleanedMountPoint)
		return nil, errMountTimeout

	case err := <-serverr:
		fuse.Unmount(cleanedMountPoint)
		log.Warn("swarmfs error serving over FUSE", "mountpoint", cleanedMountPoint, "err", err)
		return nil, err

	case <-fconn.Ready:
		log.Info("swarmfs now served over FUSE", "manifest", mhash, "mountpoint", cleanedMountPoint)
	}

	swarmfs.activeMounts[cleanedMountPoint] = mi
	return mi, nil
}

func (swarmfs *SwarmFS) Unmount(mountpoint string) (*MountInfo, error) {
	swarmfs.swarmFsLock.Lock()
	defer swarmfs.swarmFsLock.Unlock()

	cleanedMountPoint, err := filepath.Abs(filepath.Clean(mountpoint))
	if err != nil {
		return nil, err
	}

	mountInfo := swarmfs.activeMounts[cleanedMountPoint]

	if mountInfo == nil || mountInfo.MountPoint != cleanedMountPoint {
		return nil, fmt.Errorf("swarmfs %s is not mounted", cleanedMountPoint)
	}
	err = fuse.Unmount(cleanedMountPoint)
	if err != nil {
		err1 := externalUnmount(cleanedMountPoint)
		if err1 != nil {
			errStr := fmt.Sprintf("swarmfs unmount error: %v", err)
			log.Warn(errStr)
			return nil, err1
		}
	}

	mountInfo.fuseConnection.Close()
	delete(swarmfs.activeMounts, cleanedMountPoint)

	succString := fmt.Sprintf("swarmfs unmounting %v succeeded", cleanedMountPoint)
	log.Info(succString)

	return mountInfo, nil
}

func (swarmfs *SwarmFS) Listmounts() []*MountInfo {
	swarmfs.swarmFsLock.RLock()
	defer swarmfs.swarmFsLock.RUnlock()

	rows := make([]*MountInfo, 0, len(swarmfs.activeMounts))
	for _, mi := range swarmfs.activeMounts {
		rows = append(rows, mi)
	}
	return rows
}

func (swarmfs *SwarmFS) Stop() bool {
	for mp := range swarmfs.activeMounts {
		mountInfo := swarmfs.activeMounts[mp]
		swarmfs.Unmount(mountInfo.MountPoint)
	}
	return true
}
