package sort

import (
    "github.com/sirupsen/logrus"
    "testing"
)

// https://zh.mojotv.cn/algorithm/golang-quick-sort
// https://juejin.cn/post/7174987674858553405

/**
从数列中挑出一个元素,称为 “基准”(pivot)
重新排序数列,所有元素比基准值小的摆放在基准前面,所有元素比基准值大的摆在基准的后面(相同的数可以到任一边);
在这个分区退出之后,该基准就处于数列的中间位置.这个称为分区(partition)操作;
递归地(recursive)把小于基准值元素的子数列和大于基准值元素的子数列排序;
*/
func QuickSort(list []int, low, high int) {
    if high > low {
        //位置划分
        pivot := partition(list, low, high)
        //左边部分排序
        QuickSort(list, low, pivot-1)
        //右边排序
        QuickSort(list, pivot+1, high)
    }
}
func partition(list []int, low, high int) int {
    pivot := list[low] //导致 low 位置值为空
    for low < high {
        //high指针值 >= pivot high指针👈移
        for low < high && pivot <= list[high] {
            high--
        }
        //填补low位置空值
        //high指针值 < pivot high值 移到low位置
        //high 位置值空
        list[low] = list[high]
        //low指针值 <= pivot low指针👉移
        for low < high && pivot >= list[low] {
            low++
        }
        //填补high位置空值
        //low指针值 > pivot low值 移到high位置
        //low位置值空
        list[high] = list[low]
    }
    //pivot 填补 low位置的空值
    list[low] = pivot
    return low
}

// NlogN
func TestQuickSort(test *testing.T) {
    arr := []int{20, 7, 3, 10, 15, 25, 30, 17, 19}

    QuickSort(arr, 0, len(arr)-1)
    logrus.Infof("%+v", arr)
}

// 冒泡排序，O(n^2)
func TestBubbleSort(test *testing.T) {
    arr := []int{20, 7, 3, 10, 15, 25, 30, 17, 19}
    k := len(arr)
    for i := 0; i < k-1; i++ {
        for j := 0; j < k-1-i; j++ {
            if arr[j] > arr[j+1] {
                arr[j], arr[j+1] = arr[j+1], arr[j]
            }
        }
    }

    logrus.Infof("%v", arr)
}
