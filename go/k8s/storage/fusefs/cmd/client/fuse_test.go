package client

import (
	"fmt"
	"k8s-lx1036/k8s/storage/fuse/fuseops"
	"k8s-lx1036/k8s/storage/fusefs/cmd/client/meta"
	"os"
	"testing"

	"k8s.io/klog/v2"
)

const (
	VolumeName = "pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0"
	MasterAddr = "10.160.161.13:9500,10.160.161.14:9500,10.160.161.31:9500"
)

// INFO:
//  ./globalmount 下有 1.txt 和 2.txt 两个文件, 在该 volume pvc-30e5e4f1-e04f-4d53-9c76-50a6bc35abb0 下的两个文件 inode 是不一样的，是由 meta cluster 分配的
//  [LookUpInode]inodeID:43909, name:2.txt, 对应着 partitionID=266, leaderAddr="100.160.161.50:9021"
//  [LookUpInode]inodeID:33598233, name:1.txt，对应着 partitionID=268, leaderAddr="100.160.161.50:9021"

// 最新数据：
// 1.txt
// 2.txt
// 3.txt
// abc/1.txt

func getMetaClient() *meta.MetaClient {
	client, _ := meta.NewMetaClient(VolumeName, MasterAddr)
	return client
}

func TestFuseFS(test *testing.T) {
	client := getMetaClient()
	total, used, _ := client.Statfs()
	klog.Infof(fmt.Sprintf("volume %s stat: total %dMB, used %dB", VolumeName, total>>20, used))
}

func TestGetInodeInfoByInodeID(test *testing.T) {
	client := getMetaClient()
	inodeID := fuseops.InodeID(43909) // "2.txt"
	//inodeID := fuseops.InodeID(33598233) // "1.txt"

	// {Inode:43909 Mode:420 Nlink:1 Size:5 Uid:0 Gid:0 Generation:1 ModifyTime:1648103772 CreateTime:1648103765 AccessTime:1648103772 Target:[] PInode:1}
	// {Inode:33598233 Mode:420 Nlink:1 Size:4 Uid:0 Gid:0 Generation:1 ModifyTime:1648021784 CreateTime:1648021784 AccessTime:1648021784 Target:[] PInode:1}
	_, err := client.GetInode(inodeID)
	if err != nil {
		klog.Error(err)
		return
	}
}

func TestLookUpInodeFromParentInodeByChildName(test *testing.T) {
	client := getMetaClient()
	parentInode := fuseops.InodeID(1)
	name := "2.txt"
	childInodeID, mode, err := client.Lookup(parentInode, name)
	if err != nil {
		klog.Error(err)
		return
	}

	// inode:43909, mode:-rw-r--r-- for name:2.txt
	klog.Infof(fmt.Sprintf("inode:%d, mode:%s for name:%s", childInodeID, os.FileMode(mode).String(), name))
}

func TestReadDir(test *testing.T) {
	client := getMetaClient()
	//parentInode := fuseops.InodeID(1)
	parentInode := fuseops.InodeID(16821113) // "abc" dir
	children, err := client.ReadDir(parentInode)
	if err != nil {
		klog.Error(err)
		return
	}

	// readDir 根目录下
	// {Name:1.txt Inode:33598233 Type:420}
	// {Name:2.txt Inode:43909 Type:420}
	// {Name:3.txt Inode:43910 Type:420}
	// {Name:abc Inode:16821113 Type:2147484141}

	// readDir abc 目录下
	// {Name:1.txt Inode:33598380 Type:420}
	for _, dentry := range children {
		klog.Infof(fmt.Sprintf("%+v", dentry))
	}
}
