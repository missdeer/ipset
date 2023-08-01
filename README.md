# ipset
ipset plugin for CoreDNS

[中文简要使用说明](https://github.com/missdeer/ipset/issues/4#issuecomment-974837810)，感谢[@ioiioo](https://github.com/ioiioo)。

## Example

```js
.:53 {
        forward . 192.168.1.1
        ipset {
                twiroute api.twitter.com twitter.com dev.twitter.com www.twitter.com
                route2 domain1.net domain2.net domain3.net
        }
        ipset testroute1 testroute2 testroute3
        log
}
```