package sort

import (
    "github.com/sirupsen/logrus"
    "testing"
)

// https://www.kancloud.cn/digest/batu-go/153531
// https://zh.mojotv.cn/algorithm/golang-quick-sort
// https://juejin.cn/post/7174987674858553405

func QuickSort(arr []int, left, right int) int {
    key := left     //取最左边的为key
    fast := key + 1 //快指针
    slow := key     //慢指针
    for fast <= right {
        if arr[fast] < arr[key] { //当快指针指向元素小于key就交换
            arr[slow], arr[fast] = arr[fast], arr[slow]
            slow++
        }
        fast++
    }
    arr[key], arr[slow] = arr[slow], arr[key] //慢指针回退一位再交换
    return slow                               //返回key的位置
}

func TestQuickSort(test *testing.T) {
    arr := []int{20, 7, 3, 10, 15, 25, 30, 17, 19}

    QuickSort(arr, arr[0], arr[len(arr)-1])

    logrus.Infof("%+v", arr)
}
