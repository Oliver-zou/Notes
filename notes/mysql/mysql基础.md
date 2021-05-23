### 1. [JOIN](http://www.codeproject.com/Articles/33052/Visual-Representation-of-SQL-Joins)

**`1.INNER JOIN（内连接）`**

<div align="center"> <img src="../../pics/84a7a2bf45cfe52b2623cffc59e853a8.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list> 
FROM Table_A A
INNER JOIN Table_B B
ON A.Key = B.Key
```

**`2.LEFT JOIN（左连接）`**

<div align="center"> <img src="../../pics/a44ab796c630c5e9390dafd5c0eec99c.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list>
FROM Table_A A
LEFT JOIN Table_B B
ON A.Key = B.Key
```

**`3.RIGHT JOIN（右连接）`**

<div align="center"> <img src="../../pics/2debd59ca3c68ce754dab7e497e9467b.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list>
FROM Table_A A
RIGHT JOIN Table_B B
ON A.Key = B.Key
```

**`4.OUTER JOIN（外连接）`**

<div align="center"> <img src="../../pics/c1a482cfb5b058061c219a8665d0e2c7.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list>
FROM Table_A A
FULL OUTER JOIN Table_B B
ON A.Key = B.Key
```

**`5.LEFT JOIN EXCLUDING INNER JOIN（左连接-内连接）`**

<div align="center"> <img src="../../pics/9cf5bbfb368cc5177c2e7add676f595d.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list> 
FROM Table_A A
LEFT JOIN Table_B B
ON A.Key = B.Key
WHERE B.Key IS NULL
```

**`6.RIGHT JOIN EXCLUDING INNER JOIN（右连接-内连接）`**

<div align="center"> <img src="../../pics/63b8895e3f9f17a7f93219506269e562.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list>
FROM Table_A A
RIGHT JOIN Table_B B
ON A.Key = B.Key
WHERE A.Key IS NULL
```

**`7.OUTER JOIN EXCLUDING INNER JOIN（外连接-内连接）`**

<div align="center"> <img src="../../pics/6f6882d7434065e6e8fbb35531ebc35a.jpg" width="300px"/> </div><br>

```sql
SELECT <select_list>
FROM Table_A A
FULL OUTER JOIN Table_B B
ON A.Key = B.Key
WHERE A.Key IS NULL OR B.Key IS NULL
```
