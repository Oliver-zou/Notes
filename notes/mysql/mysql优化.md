

## 一、SQL慢查询的原因

<div align="center"> <img src="../../pics/20210526172308.png" width="500px"/> </div><br>

## 1 写操作

**刷脏页**

脏页的定义是这样的：内存数据页和磁盘数据页不一致时，那么称这个内存数据页为脏页。

那为什么会出现脏页，刷脏页又怎么会导致 SQL 变慢呢？那就需要我们来看看写操作时的流程是什么样的。

对于一条写操作的 SQL 来说，执行的过程中涉及到写日志，内存及同步磁盘这几种情况。

<div align="center"> <img src="../../pics/20210526172439.png" width="500px"/> </div><br>

这里要提到一个日志文件，那就是 redo log，位于存储引擎层，用来存储物理日志。在写操作的时候，存储引擎（这里讨论的是 Innodb）会将记录写入到 redo log 中，并更新缓存，这样更新操作就算完成了。后续操作存储引擎会在适当的时候把操作记录同步到磁盘里。

看到这里你可能会有个疑问，redo log 不是日志文件吗，日志文件就存储在磁盘上，那写的时候岂不很慢吗？

其实，写redo log 的过程是顺序写磁盘的，磁盘顺序写减少了寻道等时间，速度比随机写要快很多（ 类似Kafka存储原理），因此写 redo log 速度是很快的。

好了，让我们回到开始时候的问题，为什么会出现脏页，并且脏页为什么会使 SQL 变慢。你想想，redo log 大小是一定的，且是循环写入的。在高并发场景下，redo log 很快被写满了，但是数据来不及同步到磁盘里，这时候就会产生脏页，并且还会阻塞后续的写入操作。SQL 执行自然会变慢。

**锁**

写操作时 SQL 慢的另一种情况是可能遇到了锁，这个很容易理解。举个例子，你和别人合租了一间屋子，只有一个卫生间，你们俩同时都想去，但对方比你早了一丢丢。那么此时你只能等对方出来后才能进去。

对应到 Mysql 中，当某一条 SQL 所要更改的行刚好被加了锁，那么此时只有等锁释放了后才能进行后续操作。

但是还有一种极端情况，你的室友一直占用着卫生间，那么此时你该怎么整，总不能尿裤子吧，多丢人。对应到Mysql 里就是遇到了死锁或是锁等待的情况。这时候该如何处理呢？

Mysql 中提供了查看当前锁情况的方式：

```sql
select * from information_schema.INNODB_TRX;
```

通过在命令行执行图中的语句，可以查看当前运行的事务情况，这里介绍几个查询结果中重要的参数：

<div align="center"> <img src="../../pics/20210526172645.png" width="500px"/> </div><br>

当前事务如果等待时间过长或出现死锁的情况，可以通过 「**kill 线程ID**」 的方式释放当前的锁。

这里的线程 ID 指表中 **trx_mysql_thread_id** 参数。

## 2. 读操作

说完了写操作，读操作大家可能相对来说更熟悉一些。SQL 慢导致读操作变慢的问题在工作中是经常会被涉及到的。

在讲读操作变慢的原因之前我们先来看看是如何定位慢 SQL 的。Mysql 中有一个叫作**慢查询日志**的东西，它是用来记录超过指定时间的 SQL 语句的。默认情况下是关闭的，通过手动配置才能开启慢查询日志进行定位。
具体的配置方式是这样的：

- 查看当前慢查询日志的开启情况：

<div align="center"> <img src="../../pics/20210526174000.png" width="500px"/> </div><br>

- 开启慢查询日志（临时）：

  ```sql
  set global slow_query_log='ON';
  ```

  <div align="center"> <img src="../../pics/20210526174228.png" width="500px"/> </div><br>

  注意这里只是临时开启了慢查询日志，如果 mysql 重启后则会失效。可以 my.cnf 中进行配置使其永久生效。

**存在原因**

**（1）未命中索引**

SQL 查询慢的原因之一是可能未命中索引，关于使用索引为什么能使查询变快以及使用时的注意事项，见下。

**（2）脏页问题**

另一种还是我们上边所提到的刷脏页情况，只不过和写操作不同的是，是在读时候进行刷脏页的。

是不是有点懵逼，别急，听我娓娓道来：

为了避免每次在读写数据时访问磁盘增加 IO 开销，Innodb 存储引擎通过把相应的数据页和索引页加载到内存的缓冲池（buffer pool）中来提高读写速度。然后按照最近最少使用原则来保留缓冲池中的缓存数据。

那么当要读入的数据页不在内存中时，就需要到缓冲池中申请一个数据页，但缓冲池中数据页是一定的，当数据页达到上限时此时就需要把最久不使用的数据页从内存中淘汰掉。但如果淘汰的是脏页呢，那么就需要把脏页刷到磁盘里才能进行复用。你看，又回到了刷脏页的情况，读操作时变慢你也能理解了吧？

**防患于未然**

知道了原因，我们如何来避免或缓解这种情况呢？

首先来看未命中索引的情况：

不知道大家有没有使用 Mysql 中 explain 的习惯，反正我是每次都会用它来查看下当前 SQL 命中索引的情况。避免其带来一些未知的隐患。这里简单介绍下其使用方式，通过在所执行的 SQL 前加上 explain 就可以来分析当前 SQL 的执行计划：

<div align="center"> <img src="../../pics/20210526174440.png" width="500px"/> </div><br>

执行后的结果对应的字段概要描述如下图所示：

<div align="center"> <img src="../../pics/20210526174504.jpg" width="500px"/> </div><br>

这里需要重点关注以下几个字段：

**1、type**

表示 MySQL 在表中找到所需行的方式。其中常用的类型有：ALL、index、range、 ref、eq_ref、const、system、NULL 这些类型从左到右，性能逐渐变好。

- ALL：Mysql 遍历全表来找到匹配的行；
- index：与 ALL 区别为 index 类型只遍历索引树；
- range：只检索给定范围的行，使用一个索引来选择行；
- ref：表示上述表的连接匹配条件，哪些列或常量被用于查找索引列上的值；
- eq_ref：类似ref，区别在于使用的是否为唯一索引。对于每个索引键值，表中只有一条记录匹配，简单来说，就是多表连接中使用 primary key 或者 unique key作为关联条件；
- const、system：当 Mysql 对查询某部分进行优化，并转换为一个常量时，使用这些类型访问。如将主键置于 where 列表中，Mysql 就能将该查询转换为一个常量，system 是 const类型的特例，当查询的表只有一行的情况下，使用system；
- NULL：Mysql 在优化过程中分解语句，执行时甚至不用访问表或索引，例如从一个索引列里选取最小值可以通过单独索引查找完成。

**2、possible_keys**：查询时可能使用到的索引（但不一定会被使用，没有任何索引时显示为 NULL）。

**3、key**：实际使用到的索引。

**4、rows**：估算查找到对应的记录所需要的行数。

**5、Extra**

比较常见的是下面几种：

- Useing index：表明使用了覆盖索引，无需进行回表；
- Using where：不用读取表中所有信息，仅通过索引就可以获取所需数据，这发生在对表的全部的请求列都是同一个索引的部分的时候，表示mysql服务器将在存储引擎检索行后再进行过滤；
- Using temporary：表示MySQL需要使用临时表来存储结果集，常见于排序和分组查询，常见 group by，order by；
- Using filesort：当Query中包含 order by 操作，而且无法利用索引完成的排序操作称为“文件排序”。

对于刷脏页的情况，我们需要控制脏页的比例，不要让它经常接近 75%。同时还要控制 redo log 的写盘速度，并且通过设置 innodb_io_capacity 参数告诉 InnoDB 你的磁盘能力。

**总结**

**写操作**

- 当 redo log 写满时就会进行刷脏页，此时写操作也会终止，那么 SQL 执行自然就会变慢。
- 遇到所要修改的数据行或表加了锁时，需要等待锁释放后才能进行后续操作，SQL 执行也会变慢。

**读操作**

- 读操作慢很常见的原因是未命中索引从而导致全表扫描，可以通过 explain 方式对 SQL 语句进行分析。
- 另一种原因是在读操作时，要读入的数据页不在内存中，需要通过淘汰脏页才能申请新的数据页从而导致执行变慢。

## 二、案例分析：海量数据场景优化

### 1. 准备表数据

建一张用户表，表中的字段有用户ID、用户名、地址、记录创建时间，写一个存储过程插入一百万条数据。

```sql
CREATE TABLE `t_user` (
  `id` int NOT NULL,
  `user_name` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT NULL,
  `address` varchar(255) DEFAULT NULL,
  `create_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DELIMITER ;;
CREATE PROCEDURE user_insert()
BEGIN
DECLARE i INT DEFAULT 0;
WHILE i<1000000
DO
INSERT INTO t_user(id, user_name, address,  create_time) VALUES (i, CONCAT('mayun',i), '浙江杭州', now());
SET i=i+1;
END WHILE ;
commit;
END;;
CALL user_insert();
```

### 2. 优化

### 2.1 SQL查询优化

#### 1. 分页

##### 1.1 使用自增id代替offset

**1.OFFSET 和 LIMIT 有什么问题？**

OFFSET 和 LIMIT 对于数据量少的项目来说是没问题的。但当数据库中的数据量超过服务器内存能够存储的能力，并需要对所有数据进行分页，就会出现问题。为了实现分页，每次收到分页请求时，数据库都需要进行低效的全表扫描。

> 什么是全表扫描？全表扫描 (又称顺序扫描) 就是在数据库中进行逐行扫描，顺序读取表中的每一行记录，然后检查各个列是否符合查询条件。这种扫描是已知最慢的，因为需要进行大量的磁盘 I/O，而且从磁盘到内存的传输开销也很大。

这意味着，如果你有 1 亿个用户，OFFSET 是 5 千万，那它需要获取所有这些记录 (包括那么多根本不需要的数据)，将它们放入内存，然后获取 LIMIT 指定的 20 条结果。也就是说，为了获取一页的数据：10万行中的第5万行到第5万零20行需要先获取 5 万行。这么做是多么低效？可以看看这个[例子](https://www.db-fiddle.com/f/3JSpBxVgcqL3W2AzfRNCyq/1?ref=hackernoon.com)：左边的 Schema SQL 将插入 10 万行数据，右边有一个性能很差的查询和一个较好的解决方案。只需单击顶部的 Run，就可以比较它们的执行时间。第一个查询的运行时间至少是第二个查询的 30 倍。

数据越多，情况就越糟。看看我对 10 万行数据进行的 [PoC](https://github.com/IvoPereira/Efficient-Pagination-SQL-PoC?ref=hackernoon.com)。因此：OFFSET 越高，查询时间就越长。

**2.替代方案**

```sql
select * from table_name where id > ? limit ?;
```

这是一种基于指针的分页。在本地保存上一次接收到的主键 (通常是一个 ID) 和 LIMIT，而不是 OFFSET 和 LIMIT，那么每一次的查询可能都与此类似。为什么？因为通过显式告知数据库最新行，数据库就确切地知道从哪里开始搜索（基于有效的索引），而不需要考虑目标范围之外的记录。

要使用这种基于游标的分页，需要有一个惟一的序列字段 (或多个)，比如惟一的整数 ID 或时间戳，但在某些特定情况下可能无法满足这个条件。建议是，不管怎样都要考虑每种解决方案的优缺点，以及需要执行哪种查询。

如果需要基于大量数据做查询操作，Rick James 的[文章](http://mysql.rjweb.org/doc.php/lists)提供了更深入的指导。

如果我们的表没有主键，比如是具有多对多关系的表，那么就使用传统的 OFFSET/LIMIT 方式，只是这样做存在潜在的慢查询问题。我建议在需要分页的表中使用自动递增的主键，即使只是为了分页。

##### 1.2 案例

- limit 1000时

<div align="center"> <img src="../../pics/20210526161206.png" width="300px"/> </div><br>

- limit 1000000时

<div align="center"> <img src="../../pics/20210526161412.png" width="300px"/> </div><br>

可以看到limit值越大，耗时越长.

- 子查询优化

<div align="center"> <img src="../../pics/20210526161732.png" width="500px"/> </div><br>

可以看到比起之前 limit 1000000时的0.218s 效率提高了很多

- 使用JOIN分页

<div align="center"> <img src="../../pics/20210526161842.png" width="500px"/> </div><br>

可以看到比起之前 limit 1000000时的0.218s 效率也同样提高了很多

- 使用前一次查询的最大ID

<div align="center"> <img src="../../pics/20210526161911.png" width="500px"/> </div><br>

可以看到这种方法效率最高，但依赖于需要知道最大ID，这种适合点击下一页查询（类似于滚动加载数据）的场景

- 通过伪列对ID进行分页

<div align="center"> <img src="../../pics/20210526161942.png" width="500px"/> </div><br>

然后可以开启多个线程去进行最高效率查询语句的批量查询操作 0~10000，10001-20000.... 这样子的话可以快速把全量数据查询出来同步至缓存中。

**分页优化总结：** 使用前一次查询的最大ID进行查询优化是效率最高的方法，但这种方法只适用于下一页点击的这种操作，对于同步全量数据来说建议的方式使用伪列对ID进行分页，然后开启多个线程同时查询，把全量数据加载到缓存，以后面试官问你如何 **快速获取海量数据并加载到缓存** 你该知道怎么回答了吧。

#### 2. SQL语句中IN包含的值不应过多

MySQL对于IN做了相应的优化，即将IN中的常量全部存储在一个数组里面，而且这个数组是排好序的。但是如果数值较多，产生的消耗也是比较大的。再例如：`select id from table_name where num in(1,2,3)` 对于连续的数值，能用 between 就不要用 `in` 了；再或者使用连接来替换。

#### 3. SELECT语句务必指明字段名称

SELECT *增加很多不必要的消耗（cpu、io、内存、网络带宽）

增加了使用覆盖索引的可能性；

当表结构发生改变时，前端也需要更新。

#### 4. 当只需要一条数据的时候，使用limit 1

这是为了使EXPLAIN中type列达到const类型

#### 5. 如果排序字段没有用到索引，就尽量少排序

#### 6. 如果限制条件中其他字段没有索引，尽量少用or

or两边的字段中，如果有一个不是索引字段，而其他条件也不是索引字段，会造成该查询不走索引的情况。很多时候使用 union all 或者是union(必要的时候)的方式来代替“or”会得到更好的效果。

<div align="center"> <img src="../../pics/20210526162702.png" width="500px"/> </div><br>

可以看到这条语句没有使用到索引，是因为当or左右查询字段只有一个是索引，该索引失效，只有当or左右查询字段均为索引时，才会生效。

#### 7. 尽量用union all代替union

`union`和`union all`的差异主要是前者需要将结果集合并后再进行唯一性过滤操作，这就会涉及到排序，增加大量的CPU运算，加大资源消耗及延迟。当然，`union all`的前提条件是两个结果集没有重复数据。

#### 8. 区分in和exists， not in和not exists

```sql
select * from 表A where id in (select id from 表B)
```

上面sql语句相当于

```sql
select * from 表A where exists
(select * from 表B where 表B.id=表A.id)
```

区分in和exists主要是造成了驱动顺序的改变（这是性能变化的关键），如果是exists，那么以外层表为驱动表，先被访问，如果是IN，那么先执行子查询。所以IN适合于外表大而内表小的情况；EXISTS适合于外表小而内表大的情况。

> 关于not in和not exists，推荐使用not exists，不仅仅是效率问题，not in可能存在逻辑问题。如何高效的写出一个替代not exists的sql语句？

原sql语句

```sql
select colname … from A表 
where a.id not in (select b.id from B表)
```

高效的sql语句

```sql
select colname … from A表 Left join B表 on 
where a.id = b.id where b.id is null
```

取出的结果集如下图表示，A表不在B表中的数据

<div align="center"> <img src="../../pics/aa6481633dba4e36f4a308adffc44e2d.jpg" width="300px"/> </div><br>

#### 9. 不使用ORDER BY RAND()

```sql
select id from `table_name` 
order by rand() limit 1000;
```

上面的sql语句，可优化为

```sql
select id from `table_name` t1 join 
(select rand() * (select max(id) from `table_name`) as nid) t2 
on t1.id > t2.nid limit 1000;
```

#### 10. 分段查询

在一些用户选择页面中，可能一些用户选择的时间范围过大，造成查询缓慢。主要的原因是扫描行数过多。这个时候可以通过程序，分段进行查询，循环遍历，将结果合并处理进行展示。

如下图这个sql语句，扫描的行数成百万级以上的时候就可以使用分段查询

<div align="center"> <img src="../../pics/1a5b22cd474483be171201ae07c48a12.jpg" width="800px"/> </div><br>

#### 11. 避免在 where 子句中对字段进行 null 值判断

对于null的判断会导致引擎放弃使用索引而进行全表扫描。

#### 12. 不建议使用%前缀模糊查询

例如`LIKE “%name”`或者`LIKE “%name%”`，这种查询会导致索引失效而进行全表扫描。

<div align="center"> <img src="../../pics/20210526162831.png" width="500px"/> </div><br>

但是可以使用`LIKE “name%”`。那如何查询`%name%`？

如下图所示，虽然给secret字段添加了索引，但在explain结果果并没有使用

<div align="center"> <img src="../../pics/a19f04f630e678a6171478ed068f274b.jpg" width="800px"/> </div><br>

那么如何解决这个问题呢，答案：使用全文索引

在我们查询中经常会用到`select id,fnum,fdst from table_name where user_name like '%zhangsan%';`。这样的语句，普通索引是无法满足查询需求的。庆幸的是在MySQL中，有全文索引来帮助我们。

创建全文索引的sql语法是：

```sql
ALTER TABLE `table_name` ADD FULLTEXT INDEX `idx_user_name` (`user_name`);
```

使用全文索引的sql语句是：

```sql
select id,fnum,fdst from table_name 
where match(user_name) against('zhangsan' in boolean mode);
```

> 注意：在需要创建全文索引之前，请联系DBA确定能否创建。同时需要注意的是查询语句的写法与普通索引的区别

#### 13. 避免在where子句中对字段进行表达式操作

比如

```
select user_id,user_project from table_name where age*2=36;
```

中对字段就行了算术运算，这会造成引擎放弃使用索引，建议改成

```
select user_id,user_project from table_name where age=36/2;
```

#### 14. 避免隐式类型转换

where 子句中出现 column 字段的类型和传入的参数类型不一致的时候发生的类型转换，建议先确定where中的参数类型

<div align="center"> <img src="../../pics/ac5959037d49620dae8b78d6cda64b31.jpg" width="800px"/> </div><br>

#### 15. 对于联合索引来说，要遵守最左前缀法则

举列来说索引含有字段`id`,`name`,`school`，可以直接用id字段，也可以`id`,`name`这样的顺序，但是`name`，`school`都无法使用这个索引。所以在创建联合索引的时候一定要注意索引字段顺序，常用的查询字段放在最前面。

<div align="center"> <img src="../../pics/20210526162924.png" width="500px"/> </div><br>

ref:这个连接类型只有在查询使用了不是惟一或主键的键或者是这些类型的部分（比如，利用最左边前缀）时发生。没有值说明没有利用最左前缀原则

再来看个使用了最左前缀的例子

<div align="center"> <img src="../../pics/20210526162953.png" width="500px"/> </div><br>

#### 16. 必要时可以使用force index来强制查询走某个索引

有的时候MySQL优化器采取它认为合适的索引来检索sql语句，但是可能它所采用的索引并不是我们想要的。这时就可以采用force index来强制优化器使用我们制定的索引。

#### 17. 注意范围查询语句

对于联合索引来说，如果存在范围查询，比如`between not < > !=`等条件时，会造成后面的索引字段失效。

#### 18. 分解关联查询 例如这条语句

<div align="center"> <img src="../../pics/20210526163212.png" width="500px"/> </div><br>

​	可以分解成

<div align="center"> <img src="../../pics/20210526163242.png" width="500px"/> </div><br>

#### 19. 关于JOIN优化

<div align="center"> <img src="../../pics/b3c1993381b21bf1836941b5e55e5aa0.jpg" width="500px"/> </div><br>

- LEFT JOIN A表为驱动表
- INNER JOIN MySQL会自动找出那个数据少的表作用驱动表
- RIGHT JOIN B表为驱动表

> 注意：MySQL中没有full join，可以用以下方式来解决

```sql
select * from A left join B on B.name = A.name where B.name is null
 union all
select * from B;
```

**尽量使用inner join，避免left join**

参与联合查询的表至少为2张表，一般都存在大小之分。如果连接方式是inner join，在没有其他过滤条件的情况下MySQL会自动选择小表作为驱动表，但是left join在驱动表的选择上遵循的是左边驱动右边的原则，即left join左边的表名为驱动表。

**合理利用索引**

**被驱动表的索引字段作为on的限制字段。**

**利用小表去驱动大表**: 即小的数据集驱动大的数据集。如：以t_user，t_order两表为例，两表通过 t_user的id字段进行关联。

```sql
当 t_order表的数据集小于t_user表时,用 in 优化 exist,使用 in,两表执行顺序是先查 t_order 表,再查t_user表
select * from t_user where id in (select user_id from t_order)
 
当 t_user 表的数据集小于 t_order 表时，用 exist 优化 in,使用 exists,两表执行顺序是先查 t_user  表,再查 t_order  表
select * from t_user where exists (select 1 from B where t_order.user_id= t_user.id)
```

<div align="center"> <img src="../../pics/228447709b69b9e7f6b89fa3277ab880.jpg" width="300px"/> </div><br>

从原理图能够直观的看出如果能够减少驱动表的话，减少嵌套循环中的循环次数，以减少 IO总量及CPU运算的次数。

**巧用STRAIGHT_JOIN**

`inner join`是由mysql选择驱动表，但是有些特殊情况需要选择另个表作为驱动表，比如有`group by`、`order by`等`「Using filesort」`、`「Using temporary」`时。`STRAIGHT_JOIN`来强制连接顺序，在`STRAIGHT_JOIN`左边的表名就是驱动表，右边则是被驱动表。在使用`STRAIGHT_JOIN`有个前提条件是该查询是内连接，也就是`inner join`。其他链接不推荐使用`STRAIGHT_JOIN`，否则可能造成查询结果不准确。

<div align="center"> <img src="../../pics/7e6e60556594d77eb3b9c581bbdb5bbd.jpg" width="800px"/> </div><br>

这个方式有时可能减少3倍的时间。

### 2.2 普通索引优化

先来看没索引优化的情况下的查询效率

<div align="center"> <img src="../../pics/20210526162109.png" width="500px"/> </div><br>

可以看到这时没用索引的情况，用了0.305S接下来看看加了索引后的结果

- 普通索引优化

<div align="center"> <img src="../../pics/20210526162144.png" width="500px"/> </div><br>

只需要0.024S，我们可以EXPLAIN看下，可以看到使用了普通索引后查询效率明显增加。

<div align="center"> <img src="../../pics/20210526162232.png" width="500px"/> </div><br>

### 2.3 复合索引优化

复合索引什么时候用

1. 单表中查询、条件语句中具有较多个字段
2. 使用索引会影响写的效率，需要研究建立最优秀的索引

建一个复合索引

<div align="center"> <img src="../../pics/20210526162512.png" width="500px"/> </div><br>

MySQL建立复合索引时实际建立了(user_name)、（user_name,address）、(user_name,address,create_time)三个索引,我们都知道每多一个索引，都会增加写操作的开销和磁盘空间的开销，对于海量数据的表，这可是不小的开销，所以你会发现我们在这里使用复合索引一个顶三个，又能减少写操作的开销和磁盘空间的开销。

当我们select user_name,address,create_time from t_user where user_name=xx and address = xxx时，MySQL可以直接通过遍历索引取得数据，无需回表，这减少了很多的随机IO操作。所以，在真正的实际应用中，这就是覆盖索引，是复合索引中主要的提升性能的优化手段之一。

### 2.4 事务优化

首先了解下事务的隔离级别，数据库共定义了四种隔离级别：

1. Serializable：可避免脏读、不可重复读、虚读情况的发生。（串行化）
2. Repeatable read：可避免脏读、不可重复读情况的发生。（可重复读）
3. Read committed：可避免脏读情况发生（读已提交）。
4. Read uncommitted：最低级别，以上情况均无法保证。(读未提交)

可以通过 set transaction isolation level 设置事务隔离级别来提高性能

### 2.5  数据库性能优化

**开启查询缓存**

- 在解析一个查询语句前，如果查询缓存是打开的，那么MySQL会检查这个查询语句是否命中查询缓存中的数据。如果当前查询恰好命中查询缓存，在检查一次用户权限后直接返回缓存中的结果。这种情况下，查询不会被解析，也不会生成执行计划，更不会执行。MySQL将缓存存放在一个引用表（不要理解成table，可以认为是类似于HashMap的数据结构），通过一个哈希值索引，这个哈希值通过查询本身、当前要查询的数据库、客户端协议版本号等一些可能影响结果的信息计算得来。所以两个查询在任何字符上的不同（例如：空格、注释），都会导致缓存不会命中。
- 如果查询中包含任何用户自定义函数、存储函数、用户变量、临时表、mysql库中的系统表，其查询结果都不会被缓存。比如函数NOW()或者CURRENT_DATE()会因为不同的查询时间，返回不同的查询结果，再比如包含CURRENT_USER或者CONNECION_ID()的查询语句会因为不同的用户而返回不同的结果，将这样的查询结果缓存起来没有任何的意义。
- 既然是缓存，就会失效，那查询缓存何时失效呢？MySQL的查询缓存系统会跟踪查询中涉及的每个表，如果这些表（数据或结构）发生变化，那么和这张表相关的所有缓存数据都将失效。正因为如此，在任何的写操作时，MySQL必须将对应表的所有缓存都设置为失效。如果查询缓存非常大或者碎片很多，这个操作就可能带来很大的系统消耗，甚至导致系统僵死一会儿。而且查询缓存对系统的额外消耗也不仅仅在写操作，读操作也不例外：
- 任何的查询语句在开始之前都必须经过检查，即使这条SQL语句永远不会命中缓存 　
- 如果查询结果可以被缓存，那么执行完成后，会将结果存入缓存，也会带来额外的系统消耗 
- 基于此，我们要知道并不是什么情况下查询缓存都会提高系统性能，缓存和失效都会带来额外消耗，只有当缓存带来的资源节约大于其本身消耗的资源时，才会给系统带来性能提升。但要如何评估打开缓存是否能够带来性能提升是一件非常困难的事情，也不在本文讨论的范畴内。如果系统确实存在一些性能问题，可以尝试打开查询缓存，并在数据库设计上做一些优化，比如：
- 批量插入代替循环单条插入  . 合理控制缓存空间大小，一般来说其大小设置为几十兆比较合适  . 可以通过SQL\_CACHE和SQL\_NO\_CACHE来控制某个查询语句是否需要进行缓存  最后的忠告是不要轻易打开查询缓存，特别是写密集型应用。如果你实在是忍不住，可以将query\_cache\_type设置为DEMAND，这时只有加入SQL\_CACHE的查询才会走缓存，其他查询则不会，这样可以非常自由地控制哪些查询需要被缓存。  当然查询缓存系统本身是非常复杂的，这里讨论的也只是很小的一部分，其他更深入的话题，比如：缓存是如何使用内存的？如何控制内存的碎片化？事务对查询缓存有何影响等等，读者可以自行阅读相关资料，这里权当抛砖引玉吧。  **语法解析和预处理**
-  MySQL通过关键字将SQL语句进行解析，并生成一颗对应的解析树。这个过程解析器主要通过语法规则来验证和解析。比如SQL中是否使用了错误的关键字或者关键字的顺序是否正确等等。预处理则会根据MySQL规则进一步检查解析树是否合法。比如检查要查询的数据表和数据列是否存在等等。

### 2.6 系统内核参数优化

~~~ini
```bash
#基础配置
datadir=/data/datafile
socket=/var/lib/mysql/mysql.sock
log-error=/data/log/mysqld.log
pid-file=/var/run/mysqld/mysqld.pid
character_set_server=utf8
#允许任意IP访问
bind-address = 0.0.0.0
#是否支持符号链接，即数据库或表可以存储在my.cnf中指定datadir之外的分区或目录，为0不开启
#symbolic-links=0
#支持大小写
lower_case_table_names=1
#二进制配置
server-id = 1
log-bin = /data/log/mysql-bin.log
log-bin-index =/data/log/binlog.index
log_bin_trust_function_creators=1
expire_logs_days=7
#sql_mode定义了mysql应该支持的sql语法，数据校验等
#mysql5.0以上版本支持三种sql_mode模式：ANSI、TRADITIONAL和STRICT_TRANS_TABLES。
#ANSI模式：宽松模式，对插入数据进行校验，如果不符合定义类型或长度，对数据类型调整或截断保存，报warning警告。
#TRADITIONAL模式：严格模式，当向mysql数据库插入数据时，进行数据的严格校验，保证错误数据不能插入，报error错误。用于事物时，会进行事物的回滚。
#STRICT_TRANS_TABLES模式：严格模式，进行数据的严格校验，错误数据不能插入，报error错误。
sql_mode=STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION
#InnoDB存储数据字典、内部数据结构的缓冲池，16MB已经足够大了。
innodb_additional_mem_pool_size = 16M
#InnoDB用于缓存数据、索引、锁、插入缓冲、数据字典等
#如果是专用的DB服务器，且以InnoDB引擎为主的场景，通常可设置物理内存的60%
#如果是非专用DB服务器，可以先尝试设置成内存的1/4
innodb_buffer_pool_size = 4G
#InnoDB的log buffer，通常设置为 64MB 就足够了
innodb_log_buffer_size = 64M
#InnoDB redo log大小，通常设置256MB 就足够了
innodb_log_file_size = 256M
#InnoDB redo log文件组，通常设置为 2 就足够了
innodb_log_files_in_group = 2
#共享表空间:某一个数据库的所有的表数据，索引文件全部放在一个文件中，默认这个共享表空间的文件路径在data目录下。默认的文件名为:ibdata1 初始化为10M。
#独占表空间:每一个表都将会生成以独立的文件方式来进行存储，每一个表都有一个.frm表描述文件，还有一个.ibd文件。其中这个文件包括了单独一个表的数据内容以及索引内容，默认情况下它的存储位置也是在表的位置之中。
#设置参数为1启用InnoDB的独立表空间模式，便于管理
innodb_file_per_table = 1
#InnoDB共享表空间初始化大小，默认是 10MB，改成 1GB，并且自动扩展
innodb_data_file_path = ibdata1:1G:autoextend
#设置临时表空间最大4G
innodb_temp_data_file_path=ibtmp1:500M:autoextend:max:4096M
#启用InnoDB的status file，便于管理员查看以及监控
innodb_status_file = 1
#当设置为0，该模式速度最快，但不太安全，mysqld进程的崩溃会导致上一秒钟所有事务数据的丢失。
#当设置为1，该模式是最安全的，但也是最慢的一种方式。在mysqld 服务崩溃或者服务器主机crash的情况下，binary log 只有可能丢失最多一个语句或者一个事务。
#当设置为2，该模式速度较快，也比0安全，只有在操作系统崩溃或者系统断电的情况下，上一秒钟所有事务数据才可能丢失。
innodb_flush_log_at_trx_commit = 1
#设置事务隔离级别为 READ-COMMITED，提高事务效率，通常都满足事务一致性要求
#transaction_isolation = READ-COMMITTED
#max_connections：针对所有的账号所有的客户端并行连接到MYSQL服务的最大并行连接数。简单说是指MYSQL服务能够同时接受的最大并行连接数。
#max_user_connections : 针对某一个账号的所有客户端并行连接到MYSQL服务的最大并行连接数。简单说是指同一个账号能够同时连接到mysql服务的最大连接数。设置为0表示不限制。
#max_connect_errors：针对某一个IP主机连接中断与mysql服务连接的次数，如果超过这个值，这个IP主机将会阻止从这个IP主机发送出去的连接请求。遇到这种情况，需执行flush hosts。
#执行flush host或者 mysqladmin flush-hosts，其目的是为了清空host cache里的信息。可适当加大，防止频繁连接错误后，前端host被mysql拒绝掉
#在 show global 里有个系统状态Max_used_connections,它是指从这次mysql服务启动到现在，同一时刻并行连接数的最大值。它不是指当前的连接情况，而是一个比较值。如果在过去某一个时刻，MYSQL服务同时有10
00个请求连接过来，而之后再也没有出现这么大的并发请求时，则Max_used_connections=1000.请注意与show variables 里的max_user_connections的区别。#Max_used_connections / max_connections * 100% ≈ 85%
max_connections=600
max_connect_errors=1000
max_user_connections=400
#设置临时表最大值，这是每次连接都会分配，不宜设置过大 max_heap_table_size 和 tmp_table_size 要设置一样大
max_heap_table_size = 100M
tmp_table_size = 100M
#每个连接都会分配的一些排序、连接等缓冲，一般设置为 2MB 就足够了
sort_buffer_size = 2M
join_buffer_size = 2M
read_buffer_size = 2M
read_rnd_buffer_size = 2M
#建议关闭query cache，有些时候对性能反而是一种损害
query_cache_size = 0
#如果是以InnoDB引擎为主的DB，专用于MyISAM引擎的 key_buffer_size 可以设置较小，8MB 已足够
#如果是以MyISAM引擎为主，可设置较大，但不能超过4G
key_buffer_size = 8M
#设置连接超时阀值，如果前端程序采用短连接，建议缩短这2个值，如果前端程序采用长连接，可直接注释掉这两个选项，是用默认配置(8小时)
#interactive_timeout = 120
#wait_timeout = 120
#InnoDB使用后台线程处理数据页上读写I/0请求的数量,允许值的范围是1-64
#假设CPU是2颗4核的，且数据库读操作比写操作多，可设置
#innodb_read_io_threads=5
#innodb_write_io_threads=3
#通过show engine innodb status的FILE I/O选项可查看到线程分配
#设置慢查询阀值，单位为秒
long_query_time = 120
slow_query_log=1 #开启mysql慢sql的日志
log_output=table,File #日志输出会写表，也会写日志文件，为了便于程序去统计，所以最好写表
slow_query_log_file=/data/log/slow.log
##针对log_queries_not_using_indexes开启后，记录慢sql的频次、每分钟记录的条数
#log_throttle_queries_not_using_indexes = 5
##作为从库时生效,从库复制中如何有慢sql也将被记录
#log_slow_slave_statements = 1
##检查未使用到索引的sql
#log_queries_not_using_indexes = 1
#快速预热缓冲池
innodb_buffer_pool_dump_at_shutdown=1
innodb_buffer_pool_load_at_startup=1
#打印deadlock日志
innodb_print_all_deadlocks=1
~~~

这些参数可按照自己的实际服务器以及数据库的大小进行适当调整，主要起参考作用

### 2.7 表字段优化

很多系统一开始并没有考虑表字段拆分的问题，因为拆分会带来逻辑、部署、运维的各种复杂度，一般以整型值为主的表在千万级以下，字符串为主的表在五百万以下，而事实上很多时候MySQL单表的性能依然有不少优化空间，甚至能正常支撑千万级以上的数据量：

下面直接看下如何去优化字段

1. 尽量使用TINYINT、SMALLINT、MEDIUM_INT作为整数类型而非INT，如果非负则加上UNSIGNED
2. 单表不要有太多字段，建议在15以内
3. 尽量使用TIMESTAMP而非DATETIME
4. 使用枚举或整数代替字符串类型
5. VARCHAR的长度只分配真正需要的空间
6. 避免使用NULL字段，很难查询优化且占用额外索引空间
7. 用整型来存IP

### 2.8 分布式场景下常用优化手段

1. 升级硬件

Scale up，这个不多说了，根据MySQL是CPU密集型还是I/O密集型，通过提升CPU和内存、使用SSD，都能显著提升MySQL性能

1. 读写分离

也是目前常用的优化，从库读主库写，一般不要采用双主或多主引入很多复杂性，尽量采用文中的其他方案来提高性能。同时目前很多拆分的解决方案同时也兼顾考虑了读写分离

1. 使用缓存

   缓存可以发生在这些层次：

   MySQL内部：在系统内核参数优化介绍了相关设置

   数据访问层：比如MyBatis针对SQL语句做缓存，而Hibernate可以精确到单个记录，这里缓存的对象主要是持久化对象Persistence Object

   应用服务层：这里可以通过编程手段对缓存做到更精准的控制和更多的实现策略，这里缓存的对象是数据传输对象Data Transfer Object

   Web层：针对web页面做缓存

   浏览器客户端：用户端的缓存

   可以根据实际情况在一个层次或多个层次结合加入缓存。这里重点介绍下服务层的缓存实现，目前主要有两种方式：

   直写式（Write Through）：在数据写入数据库后，同时更新缓存，维持数据库与缓存的一致性。这也是当前大多数应用缓存框架如Spring Cache的工作方式。这种实现非常简单，同步好，但效率一般。

   回写式（Write Back）：当有数据要写入数据库时，只会更新缓存，然后异步批量的将缓存数据同步到数据库上。这种实现比较复杂，需要较多的应用逻辑，同时可能会产生数据库与缓存的不同步，但效率非常高。

2. 水平拆分。