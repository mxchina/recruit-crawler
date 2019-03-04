# recruit-crawler
纯手写招聘信息爬虫，如果要添加网站，添加worker，然后再main中运行。爬到的感兴趣的信息微信告警  
目前已经有了 51job 智联招聘 boss直聘 3个网站  

主要结构：saver负责将数据存储到数据库，并过滤后发出企业微信告警。  
worker是个接口，包含两个方法，fetch和parse。fetch将爬取的网页通过channel传给parse，parse将网页解析成结构化数据，然后发送给saver。  
开了很多goroutine，1000条数据只需1秒  

搜索内容变更：就是改变url，因为不同的关键字会传入不同的参数，url不一样  
告警过滤器变更：在saver包里，目前只支持集中处理，不支持特定网站传入特定func参数。

![爬取到的数据](https://github.com/mxchina/recruit-crawler/blob/master/20190305023442.png)
