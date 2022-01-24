
# BGP 状态机

| 状态 | 说明 |
| --- | --- |
| Idle | 空闲。 Idle是BGP连接的第一个状态，在空闲状态，BGP等待一个启动事件，启动事件出现后，BGP初始化资源，复位连接重试计时器（Connect-Retry），发起一条TCP连接，同时转入Connect（连接）状态。 |
| Connect | 连接。在Connect状态，BGP发起第一个TCP连接，如果连接重试计时器（Connect-Retry）超时，则重新发起TCP连接，并继续保持在Connect状态。如果TCP连接成功，转入OpenSent状态。如果TCP连接失败，转入Active状态。|
| Active | 活跃。在Active状态，BGP尝试建立TCP连接，如果连接重试计时器（Connect-Retry）超时，则退回到Connect状态。如果TCP连接成功，转入OpenSent状态。如果TCP连接失败，继续保持在Active状态，并继续发起TCP连接。|
| OpenSent | 打开消息已发送。在OpenSent状态，TCP连接已经建立，BGP已经发送了第一个Open报文，BGP等待其对等体发送Open报文并对收到的Open报文进行正确性检查。如果错误，系统发送一条出错通知消息并退回到Idle状态。如果正确，BGP开始发送Keepalive报文，并复位Keepalive计时器，开始计时，同时转入OpenConfirm状态。 |
| OpenConfirm | 打开消息确认。在OpenConfirm状态，BGP发送一个Keepalive报文，同时复位Keepalive计时器。如果收到一个Keepalive报文，转入Established阶段，BGP邻居关系建立完成。如果TCP连接中断，则退回Idle状态。 |
| Established | 已建立BGP邻居。在Established状态，表示BGP邻居关系已经建立，BGP与邻居交换Update报文，同时复位Keepalive计时器。|
| UnEstablished | 未建立BGP邻居。|

