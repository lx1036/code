package controller


// 如果没有cache和db的sync，这时手动修改了db，比如calicoctl手动修改，那cache就没法知道。
// 所以必须两边reconcile，设计很精妙，很有道理。
// 所以，Cache必须是可以同步cache和db的Cache对象。



type CalicoCache struct {

}
