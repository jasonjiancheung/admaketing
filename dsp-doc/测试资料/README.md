DSP测试资料
====================

### 模拟测试
#### 模拟竞价请求
##### 所用到命令  
```
# adserver-address:广告服务器ip地址
# adserver-port: 广告服务器监听的端口
# adx-url: 广告服务中adx的路径
# request-data: 请求内容
curl -i "http://adserver-address[:adserver-port]/adx-url" -d "`cat request-data`"
```
###### 请求数据例子
1. 灵集相关数据  
xtrader/xtrader-banner-data: 灵集banner请求数据  
xtrader/xtrader-video-data：灵集video请求数据  
xtrader/xtrader-sample-native-3-icon-logo-main-data：灵集原生广告-icon-logo-main图片都有的例子  
xtrader/xtrader-sample-native-3-main-data：灵集原生广告3张main图片的例子  

### adx提供的测试工具
#### 灵集
参见[灵集文档-沙箱联调](../adx/xtrader)
