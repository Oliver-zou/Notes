<!-- GFM-TOC -->

* [一、数据类型](#一数据类型)
    * [基本类型](#基本类型)
    * [包装类型](#包装类型)
    * [缓存池](#缓存池)
    
* [二、String](#二string)
    * [概览](#概览)
    * [不可变的好处](#不可变的好处)
    * [String, StringBuffer and StringBuilder	](#string-stringbuffer-and-stringbuilder	)
    * [String Pool](#string-pool)
    * [new String("abc")](#new-stringabc)
    
* [三、运算](#三运算)
    * [参数传递](#参数传递)
    * [float 与 double](#float-与-double)
    * [隐式类型转换](#隐式类型转换)
    * [switch](#switch)
    
* [四、关键字](#四关键字)

    * [final](#final)
    * [static](#static)

* [五、Object 通用方法](#五object-通用方法)
  
    - [构造方法](#构造方法)
    - [匿名对象](#匿名对象)
    
    * [概览](#概览)
    * [equals()](#equals)
    * [hashCode()](#hashcode)
    * [toString()](#tostring)
    * [clone()](#clone)
    
* [六、面向对象三大特征-封装](#六面向对象三大特征-封装)
  
    * [访问权限](#访问权限)
    
* [七、面向对象三大特征-继承](#七面向对象三大特征-继承)

    * [概述](#概述)
    * [super](#super)
    * [this](#this)
    * [继承的特点](#继承的特点)

* 八、[抽象类与接口](#抽象类与接口)

    - [抽象类](#抽象类)
    - [接口](#接口)

* [九、面向对象三大特征-多态](#九面向对象三大特征-多态)

    * [概述](#概述)
    * [重写与重载](#重写与重载)

* [十、反射](#十反射)

* [十一、异常](#十一异常)

* [十二、泛型](#十二泛型)

* [十三、注解](#十三注解)

* [十四、特性](#十四特性)

    * [Java 各版本的新特性](#java-各版本的新特性)
    * [Java 与 C++ 的区别](#java-与-c-的区别)
    * [JRE or JDK](#jre-or-jdk)

* [十五、内部类](#十五内部类)

* [十六、时间](#十六时间ß)

* [参考资料](#参考资料)
  <!-- GFM-TOC -->


# 一、数据类型

## 基本类型

- byte/8
- char/16
- short/16
- int/32
- float/32
- long/64
- double/64
- boolean/\~

boolean 只有两个值：true、false，可以使用 1 bit 来存储，但是具体大小没有明确规定。JVM 会在编译时期将 boolean 类型的数据转换为 int，使用 1 来表示 true，0 表示 false。JVM 支持 boolean 数组，但是是通过读写 byte 数组来实现的。

- [Primitive Data Types](https://docs.oracle.com/javase/tutorial/java/nutsandbolts/datatypes.html)
- [The Java® Virtual Machine Specification](https://docs.oracle.com/javase/specs/jvms/se8/jvms8.pdf)

## 包装类型

**概述**

Java提供了两个类型系统，基本类型与引用类型，使用基本类型在于效率，然而很多情况，会创建对象使用，因为对象可以做更多的功能，如果想要我们的基本类型像对象一样操作（泛型中使用），就可以使用基本类型对应的包装类，如下：

| 基本类型 | 对应的包装类（位于java.lang包中） |
| -------- | --------------------------------- |
| byte     | Byte                              |
| short    | Short                             |
| int      | **Integer**                       |
| long     | Long                              |
| float    | Float                             |
| double   | Double                            |
| char     | **Character**                     |
| boolean  | Boolean                           |

**装箱与拆箱**

 构造方法:

​        Integer(int value) 构造一个新分配的 Integer 对象，它表示指定的 int 值。

​        Integer(String s) 构造一个新分配的 Integer 对象，它表示 String 参数所指示的 int 值。

​            传递的字符串,必须是基本类型的字符串,否则会抛出异常 "100" 正确  "a" 抛异常

​    静态方法:

​        static Integer valueOf(int i) 返回一个表示指定的 int 值的 Integer 实例。

​        static Integer valueOf(String s) 返回保存指定的 String 的值的 Integer 对象。

拆箱:在包装类中取出基本类型的数据(包装类->基本类型的数据)

​    成员方法:

​        int intValue() 以 int 类型返回该 Integer 的值。

基本类型与对应的包装类对象之间，来回转换的过程称为”装箱“与”拆箱“：

* **装箱**：从基本类型转换为对应的包装类对象。

* **拆箱**：从包装类对象转换为对应的基本类型。

用Integer与 int为例：（看懂代码即可）

基本数值---->包装对象

~~~java
Integer i = new Integer(4);//使用构造函数函数
Integer iii = Integer.valueOf(4);//使用包装类中的valueOf方法
~~~

包装对象---->基本数值

~~~java
int num = i.intValue();
~~~

**自动装箱与自动拆箱**

由于我们经常要做基本类型与包装类之间的转换，从Java 5（JDK 1.5）开始，基本类型与包装类的装箱、拆箱动作可以自动完成。例如：

```java
Integer i = 4;//自动装箱。相当于Integer i = Integer.valueOf(4);
i = i + 5;//等号右边：将i对象转成基本数值(自动拆箱) i.intValue() + 5;
//加法运算完成后，再次装箱，把基本数值转成对象。
```

**基本类型转换为String**

基本类型->字符串(String)

​    1.基本类型的值+""  最简单的方法(工作中常用)，如：34+""

​    2.包装类的静态方法toString(参数),不是Object类的toString() 重载

​        static String toString(int i) 返回一个表示指定整数的 String 对象。

​    3.String类的静态方法valueOf(参数)

​        static String valueOf(int i) 返回 int 参数的字符串表示形式。

**String转换成对应的基本类型** 

字符串(String)->基本类型

​    使用包装类的静态方法parseXXX("字符串");

​        Integer类: static int parseInt(String s)

​        Double类: static double parseDouble(String s)

除了Character类之外，其他所有包装类都具有parseXxx静态方法可以将字符串参数转换为对应的基本类型：

- `public static byte parseByte(String s)`：将字符串参数转换为对应的byte基本类型。
- `public static short parseShort(String s)`：将字符串参数转换为对应的short基本类型。
- `public static int parseInt(String s)`：将字符串参数转换为对应的int基本类型。
- `public static long parseLong(String s)`：将字符串参数转换为对应的long基本类型。
- `public static float parseFloat(String s)`：将字符串参数转换为对应的float基本类型。
- `public static double parseDouble(String s)`：将字符串参数转换为对应的double基本类型。
- `public static boolean parseBoolean(String s)`：将字符串参数转换为对应的boolean基本类型。

代码使用（仅以Integer类的静态方法parseXxx为例）如：

```java
public class Demo18WrapperParse {
    public static void main(String[] args) {
        int num = Integer.parseInt("100");
    }
}
```

> 注意:如果字符串参数的内容无法正确转换为对应的基本类型，则会抛出`java.lang.NumberFormatException`异常。

[Autoboxing and Unboxing](https://docs.oracle.com/javase/tutorial/java/data/autoboxing.html)

## 缓存池

new Integer(123) 与 Integer.valueOf(123) 的区别在于：

- new Integer(123) 每次都会新建一个对象；
- Integer.valueOf(123) 会使用缓存池中的对象，多次调用会取得同一个对象的引用。

```java
Integer x = new Integer(123);
Integer y = new Integer(123);
System.out.println(x == y);    // false
Integer z = Integer.valueOf(123);
Integer k = Integer.valueOf(123);
System.out.println(z == k);   // true
```

valueOf() 方法的实现比较简单，就是先判断值是否在缓存池中，如果在的话就直接返回缓存池的内容。

```java
public static Integer valueOf(int i) {
    if (i >= IntegerCache.low && i <= IntegerCache.high)
        return IntegerCache.cache[i + (-IntegerCache.low)];
    return new Integer(i);
}
```

在 Java 8 中，Integer 缓存池的大小默认为 -128\~127。

```java
static final int low = -128;
static final int high;
static final Integer cache[];

static {
    // high value may be configured by property
    int h = 127;
    String integerCacheHighPropValue =
        sun.misc.VM.getSavedProperty("java.lang.Integer.IntegerCache.high");
    if (integerCacheHighPropValue != null) {
        try {
            int i = parseInt(integerCacheHighPropValue);
            i = Math.max(i, 127);
            // Maximum array size is Integer.MAX_VALUE
            h = Math.min(i, Integer.MAX_VALUE - (-low) -1);
        } catch( NumberFormatException nfe) {
            // If the property cannot be parsed into an int, ignore it.
        }
    }
    high = h;

    cache = new Integer[(high - low) + 1];
    int j = low;
    for(int k = 0; k < cache.length; k++)
        cache[k] = new Integer(j++);

    // range [-128, 127] must be interned (JLS7 5.1.7)
    assert IntegerCache.high >= 127;
}
```

编译器会在自动装箱过程调用 valueOf() 方法，因此多个值相同且值在缓存池范围内的 Integer 实例使用自动装箱来创建，那么就会引用相同的对象。

```java
Integer m = 123;
Integer n = 123;
System.out.println(m == n); // true
```

基本类型对应的缓冲池如下：

- boolean values true and false
- all byte values
- short values between -128 and 127
- int values between -128 and 127
- char in the range \u0000 to \u007F

在使用这些基本类型对应的包装类型时，如果该数值范围在缓冲池范围内，就可以直接使用缓冲池中的对象。

在 jdk 1.8 所有的数值类缓冲池中，Integer 的缓冲池 IntegerCache 很特殊，这个缓冲池的下界是 - 128，上界默认是 127，但是这个上界是可调的，在启动 jvm 的时候，通过 -XX:AutoBoxCacheMax=&lt;size&gt; 来指定这个缓冲池的大小，该选项在 JVM 初始化的时候会设定一个名为 java.lang.IntegerCache.high 系统属性，然后 IntegerCache 初始化的时候就会读取该系统属性来决定上界。

[StackOverflow : Differences between new Integer(123), Integer.valueOf(123) and just 123
](https://stackoverflow.com/questions/9030817/differences-between-new-integer123-integer-valueof123-and-just-123)

# 二、String

## 概览

java.lang.String类代表字符串。

API当中说：Java 程序中的所有字符串字面值（如 "abc" ）都作为此类的实例实现。

其实就是说：程序当中所有的双引号字符串，都是String类的对象。（就算没有new，也照样是。）

字符串的特点：
1. 字符串的内容永不可变。【重点】
2. 正是因为字符串不可改变，所以字符串是可以共享使用的。
3. 字符串效果上相当于是char[]字符数组，但是底层原理是byte[]字节数组。
4. String 被声明为 final，因此它不可被继承。(Integer 等包装类也不能被继承）

创建字符串的常见3+1种方式。

- 三种构造方法：
  public String()：创建一个空白字符串，不含有任何内容

  ```
  // 使用空参构造
  String str1 = new String(); // 小括号留空，说明字符串什么内容都没有。
  System.out.println("第1个字符串：" + str1);
  ```

  public String(char[] array)：根据字符数组的内容，来创建对应的字符串。

  ```
  // 根据字符数组创建字符串
  char[] charArray = { 'A', 'B', 'C' };
  String str2 = new String(charArray);
  System.out.println("第2个字符串：" + str2);
  ```

  public String(byte[] array)：根据字节数组的内容，来创建对应的字符串。

  ```
  // 根据字节数组创建字符串
  byte[] byteArray = { 97, 98, 99 };
  String str3 = new String(byteArray);
  System.out.println("第3个字符串：" + str3);
  ```

- 一种直接创建：

  String str = "Hello"; // 右边直接用双引号

  ```
  // 直接创建
  String str4 = "Hello";
  System.out.println("第4个字符串：" + str4);
  ```

注意：直接写上双引号，就是字符串对象。



在 Java 8 中，String 内部使用 char 数组存储数据。

```java
public final class String
    implements java.io.Serializable, Comparable<String>, CharSequence {
    /** The value is used for character storage. */
    private final char value[];
}
```

在 Java 9 之后，String 类的实现改用 byte 数组存储字符串，同时使用 `coder` 来标识使用了哪种编码。

```java
public final class String
    implements java.io.Serializable, Comparable<String>, CharSequence {
    /** The value is used for character storage. */
    private final byte[] value;

    /** The identifier of the encoding used to encode the bytes in {@code value}. */
    private final byte coder;
}
```

value 数组被声明为 final，这意味着 value 数组初始化之后就不能再引用其它数组。并且 String 内部没有改变 value 数组的方法，因此可以保证 String 不可变。

## 比较

==是进行对象的地址值比较，如果确实需要字符串的内容比较，可以使用两个方法：

public boolean equals(Object obj)：参数可以是任何对象，只有参数是一个字符串并且内容相同的才会给true；否则返回false。

注意事项：

1. 任何对象都能用Object进行接收。
2. equals方法具有对称性，也就是a.equals(b)和b.equals(a)效果一样。
3. 如果比较双方一个常量一个变量，推荐把常量字符串写在前面（防止空指针异常）。
推荐："abc".equals(str)    不推荐：str.equals("abc")

public boolean equalsIgnoreCase(String str)：忽略大小写，进行内容比较。

## 不可变的好处

**1. 可以缓存 hash 值**  

因为 String 的 hash 值经常被使用，例如 String 用做 HashMap 的 key。不可变的特性可以使得 hash 值也不可变，因此只需要进行一次计算。

**2. String Pool 的需要**  

```
字符串常量池：程序当中直接写上的双引号字符串，就在字符串常量池中。
对于基本类型来说，==是进行数值的比较。
对于引用类型来说，==是进行【地址值】的比较。
```

例子：

<div align="center"><img src="../../pics/dc3561b2-a5fa-4789-abe2-b8be7ec87df3.png"width="800px"></img></div>

如果一个 String 对象已经被创建过了，那么就会从 String Pool 中取得引用。只有 String 是不可变的，才可能使用 String Pool。

<div align="center"> <img src="https://cs-notes-1256109796.cos.ap-guangzhou.myqcloud.com/image-20191210004132894.png"/> </div><br>

**3. 安全性**  

String 经常作为参数，String 不可变性可以保证参数不可变。例如在作为网络连接参数的情况下如果 String 是可变的，那么在网络连接过程中，String 被改变，改变 String 的那一方以为现在连接的是其它主机，而实际情况却不一定是。

**4. 线程安全**  

String 不可变性天生具备线程安全，可以在多个线程中安全地使用。

[Program Creek : Why String is immutable in Java?](https://www.programcreek.com/2013/04/why-string-is-immutable-in-java/)

## String, StringBuffer and StringBuilder	

**1.字符串拼接问题**

由于String类的对象内容不可改变，所以每当进行字符串拼接时，总是会在内存中创建一个新的对象。例如：

~~~java
public class StringDemo {
    public static void main(String[] args) {
        String s = "Hello";
        s += "World";
        System.out.println(s);
    }
}
~~~

在API中对String类有这样的描述：字符串是常量，它们的值在创建后不能被更改。

根据这句话分析我们的代码，其实总共产生了三个字符串，即`"Hello"`、`"World"`和`"HelloWorld"`。引用变量s首先指向`Hello`对象，最终指向拼接出来的新字符串对象，即`HelloWord` 。

<div align="center"><img src="../../pics/91cbabd0-97ec-409d-a4a7-b94cf4948ae7.png"width="800px"></img></div>

由此可知，如果对字符串进行拼接操作，每次拼接，都会构建一个新的String对象，既耗时，又浪费空间。为了解决这一问题，可以使用`java.lang.StringBuilder`类。

查阅`java.lang.StringBuilder`的API，StringBuilder又称为可变字符序列，它是一个类似于 String 的字符串缓冲区，通过某些方法调用可以改变该序列的长度和内容。

原来StringBuilder是个字符串的缓冲区，即它是一个容器，容器中可以装很多字符串。并且能够对其中的字符串进行各种操作。

它的内部拥有一个数组用来存放字符串内容，进行字符串拼接时，直接在数组中加入新内容。StringBuilder会自动维护数组的扩容。

**2. stringbuilder的构造方法及常用方法**

常用构造方法有2个：

- `public StringBuilder()`：构造一个空的StringBuilder容器。
- `public StringBuilder(String str)`：构造一个StringBuilder容器，并将字符串添加进去。

```java
public class StringBuilderDemo {
    public static void main(String[] args) {
        StringBuilder sb1 = new StringBuilder();
        System.out.println(sb1); // (空白)
        // 使用带参构造
        StringBuilder sb2 = new StringBuilder("itcast");
        System.out.println(sb2); // itcast
    }
}
```

StringBuilder常用的方法有2个：

- `public StringBuilder append(...)`：添加任意类型数据的字符串形式，并返回当前对象自身。
- `public String toString()`：将当前StringBuilder对象转换为String对象。

**append方法**

append方法具有多种重载形式，可以接收任意类型的参数。任何数据作为参数都会将对应的字符串内容添加到StringBuilder中。例如：

```java
public class Demo02StringBuilder {
	public static void main(String[] args) {
		//创建对象
		StringBuilder builder = new StringBuilder();
		//public StringBuilder append(任意类型)
		StringBuilder builder2 = builder.append("hello");
		//对比一下
		System.out.println("builder:"+builder);
		System.out.println("builder2:"+builder2);
		System.out.println(builder == builder2); //true
	    // 可以添加 任何类型
		builder.append("hello");
		builder.append("world");
		builder.append(true);
		builder.append(100);
		// 在我们开发中，会遇到调用一个方法后，返回一个对象的情况。然后使用返回的对象继续调用方法。
        // 这种时候，我们就可以把代码现在一起，如append方法一样，代码如下
		//链式编程
		builder.append("hello").append("world").append(true).append(100);
		System.out.println("builder:"+builder);
	}
}
```

> 备注：StringBuilder已经覆盖重写了Object当中的toString方法。

**toString方法**

通过toString方法，StringBuilder对象将会转换为不可变的String对象。如：

```java
public class Demo16StringBuilder {
    public static void main(String[] args) {
        // 链式创建
        StringBuilder sb = new StringBuilder("Hello").append("World").append("Java");
        // 调用方法
        String str = sb.toString();
        System.out.println(str); // HelloWorldJava
    }
}
```

**3. 可变性**  

- String 不可变
- StringBuffer 和 StringBuilder 可变

**4. 线程安全**  

- String 不可变，因此是线程安全的
- StringBuilder 不是线程安全的
- StringBuffer 是线程安全的，内部使用 synchronized 进行同步

[StackOverflow : String, StringBuffer, and StringBuilder](https://stackoverflow.com/questions/2971315/string-stringbuffer-and-stringbuilder)

## String Pool

字符串常量池（String Pool）保存着所有字符串字面量（literal strings），这些字面量在编译时期就确定。不仅如此，还可以使用 String 的 intern() 方法在运行过程将字符串添加到 String Pool 中。

当一个字符串调用 intern() 方法时，如果 String Pool 中已经存在一个字符串和该字符串值相等（使用 equals() 方法进行确定），那么就会返回 String Pool 中字符串的引用；否则，就会在 String Pool 中添加一个新的字符串，并返回这个新字符串的引用。

下面示例中，s1 和 s2 采用 new String() 的方式新建了两个不同字符串，而 s3 和 s4 是通过 s1.intern() 和 s2.intern() 方法取得同一个字符串引用。intern() 首先把 "aaa" 放到 String Pool 中，然后返回这个字符串引用，因此 s3 和 s4 引用的是同一个字符串。

```java
String s1 = new String("aaa");
String s2 = new String("aaa");
System.out.println(s1 == s2);           // false
String s3 = s1.intern();
String s4 = s2.intern();
System.out.println(s3 == s4);           // true
```

如果是采用 "bbb" 这种字面量的形式创建字符串，会自动地将字符串放入 String Pool 中。

```java
String s5 = "bbb";
String s6 = "bbb";
System.out.println(s5 == s6);  // true
```

在 Java 7 之前，String Pool 被放在运行时常量池中，它属于永久代。而在 Java 7，String Pool 被移到堆中。这是因为永久代的空间有限，在大量使用字符串的场景下会导致 OutOfMemoryError 错误。

- [StackOverflow : What is String interning?](https://stackoverflow.com/questions/10578984/what-is-string-interning)
- [深入解析 String#intern](https://tech.meituan.com/in_depth_understanding_string_intern.html)

## new String("abc")

使用这种方式一共会创建两个字符串对象（前提是 String Pool 中还没有 "abc" 字符串对象）。

- "abc" 属于字符串字变量，因此编译时期会在 String Pool 中创建一个字符串对象，指向这个 "abc" 字符串字面量；
- 而使用 new 的方式会在堆中创建一个字符串对象。

创建一个测试类，其 main 方法中使用这种方式来创建字符串对象。

```java
public class NewStringTest {
    public static void main(String[] args) {
        String s = new String("abc");
    }
}
```

使用 javap -verbose 进行反编译，得到以下内容：

```java
// ...
Constant pool:
// ...
   #2 = Class              #18            // java/lang/String
   #3 = String             #19            // abc
// ...
  #18 = Utf8               java/lang/String
  #19 = Utf8               abc
// ...

  public static void main(java.lang.String[]);
    descriptor: ([Ljava/lang/String;)V
    flags: ACC_PUBLIC, ACC_STATIC
    Code:
      stack=3, locals=2, args_size=1
         0: new           #2                  // class java/lang/String
         3: dup
         4: ldc           #3                  // String abc
         6: invokespecial #4                  // Method java/lang/String."<init>":(Ljava/lang/String;)V
         9: astore_1
// ...
```

在 Constant Pool 中，#19 存储这字符串字面量 "abc"，#3 是 String Pool 的字符串对象，它指向 #19 这个字符串字面量。在 main 方法中，0: 行使用 new #2 在堆中创建一个字符串对象，并且使用 ldc #3 将 String Pool 中的字符串对象作为 String 构造函数的参数。

以下是 String 构造函数的源码，可以看到，在将一个字符串对象作为另一个字符串对象的构造函数参数时，并不会完全复制 value 数组内容，而是都会指向同一个 value 数组。

```java
public String(String original) {
    this.value = original.value;
    this.hash = original.hash;
}
```

# 三、运算

## 参数传递

Java 的参数是以值传递的形式传入方法中，而不是引用传递。

以下代码中 Dog dog 的 dog 是一个指针，存储的是对象的地址。在将一个参数传入一个方法时，本质上是将对象的地址以值的方式传递到形参中。

```java
public class Dog {

    String name;

    Dog(String name) {
        this.name = name;
    }

    String getName() {
        return this.name;
    }

    void setName(String name) {
        this.name = name;
    }

    String getObjectAddress() {
        return super.toString();
    }
}
```

在方法中改变对象的字段值会改变原对象该字段值，因为引用的是同一个对象。

```java
class PassByValueExample {
    public static void main(String[] args) {
        Dog dog = new Dog("A");
        func(dog);
        System.out.println(dog.getName());          // B
    }

    private static void func(Dog dog) {
        dog.setName("B");
    }
}
```

但是在方法中将指针引用了其它对象，那么此时方法里和方法外的两个指针指向了不同的对象，在一个指针改变其所指向对象的内容对另一个指针所指向的对象没有影响。

```java
public class PassByValueExample {
    public static void main(String[] args) {
        Dog dog = new Dog("A");
        System.out.println(dog.getObjectAddress()); // Dog@4554617c
        func(dog);
        System.out.println(dog.getObjectAddress()); // Dog@4554617c
        System.out.println(dog.getName());          // A
    }

    private static void func(Dog dog) {
        System.out.println(dog.getObjectAddress()); // Dog@4554617c
        dog = new Dog("B");
        System.out.println(dog.getObjectAddress()); // Dog@74a14482
        System.out.println(dog.getName());          // B
    }
}
```

[StackOverflow: Is Java “pass-by-reference” or “pass-by-value”?](https://stackoverflow.com/questions/40480/is-java-pass-by-reference-or-pass-by-value)

## float 与 double

Java 不能隐式执行向下转型，因为这会使得精度降低。

1.1 字面量属于 double 类型，不能直接将 1.1 直接赋值给 float 变量，因为这是向下转型。

```java
// float f = 1.1;
```

1.1f 字面量才是 float 类型。

```java
float f = 1.1f;
```

## 隐式类型转换

因为字面量 1 是 int 类型，它比 short 类型精度要高，因此不能隐式地将 int 类型向下转型为 short 类型。

```java
short s1 = 1;
// s1 = s1 + 1;
```

但是使用 += 或者 ++ 运算符会执行隐式类型转换。

```java
s1 += 1;
s1++;
```

上面的语句相当于将 s1 + 1 的计算结果进行了向下转型：

```java
s1 = (short) (s1 + 1);
```

[StackOverflow : Why don't Java's +=, -=, *=, /= compound assignment operators require casting?](https://stackoverflow.com/questions/8710619/why-dont-javas-compound-assignment-operators-require-casting)

## switch

从 Java 7 开始，可以在 switch 条件判断语句中使用 String 对象。

```java
String s = "a";
switch (s) {
    case "a":
        System.out.println("aaa");
        break;
    case "b":
        System.out.println("bbb");
        break;
}
```

switch 不支持 long，是因为 switch 的设计初衷是对那些只有少数几个值的类型进行等值判断，如果值过于复杂，那么还是用 if 比较合适。

```java
// long x = 111;
// switch (x) { // Incompatible types. Found: 'long', required: 'char, byte, short, int, Character, Byte, Short, Integer, String, or an enum'
//     case 111:
//         System.out.println(111);
//         break;
//     case 222:
//         System.out.println(222);
//         break;
// }
```

[StackOverflow : Why can't your switch statement data type be long, Java?](https://stackoverflow.com/questions/2676210/why-cant-your-switch-statement-data-type-be-long-java)


# 四、关键字

- 1. 

## final

final关键字代表最终、不可改变的。

常见四种用法：
1. 可以用来修饰一个类
2. 可以用来修饰一个方法
3. 还可以用来修饰一个局部变量
4. 还可以用来修饰一个成员变量

**1. 数据**  

声明数据为常量，可以是编译时常量，也可以是在运行时被初始化后不能被改变的常量。

- 对于基本类型，final 使数值不变；
- 对于引用类型，final 使引用不变，也就不能引用其它对象，但是被引用的对象本身是可以修改的。

```java
final int x = 1;
// x = 2;  // cannot assign value to final variable 'x'
final A y = new A();
y.a = 1;
```

**1.1** 局部变量

一旦使用final用来修饰局部变量，那么这个变量就不能进行更改。

“唯一一次赋值，终生不变”

对于基本类型来说，不可变说的是变量当中的数据不可改变
对于引用类型来说，不可变说的是变量当中的地址值不可改变

```java
        // 对于基本类型来说，不可变说的是变量当中的数据不可改变
        // 对于引用类型来说，不可变说的是变量当中的地址值不可改变
        Student stu1 = new Student("赵丽颖");
        System.out.println(stu1);
        System.out.println(stu1.getName()); // 赵丽颖
        stu1 = new Student("霍建华");
        System.out.println(stu1);
        System.out.println(stu1.getName()); // 霍建华
        System.out.println("===============");

        final Student stu2 = new Student("高圆圆");
        // 错误写法！final的引用类型变量，其中的地址不可改变，但内容可变
//        stu2 = new Student("赵又廷");
        System.out.println(stu2.getName()); // 高圆圆
        stu2.setName("高圆圆圆圆圆圆");
        System.out.println(stu2.getName()); // 高圆圆圆圆圆圆
```

**1.2 成员变量**

对于成员变量来说，如果使用final关键字修饰，那么这个变量也照样是不可变。

1. 由于成员变量具有默认值，所以用了final之后必须手动赋值，不会再给默认值了。
2. 对于final的成员变量，要么使用直接赋值，要么通过构造方法赋值。二者选其一。
3. 必须保证类当中所有重载的构造方法，都最终会对final的成员变量进行赋值。否则，用直接赋值。

**2. 方法**  

当final关键字用来修饰一个方法的时候，这个方法就是最终方法，也就是不能被覆盖重写。
格式：
修饰符 final 返回值类型 方法名称(参数列表) {
    // 方法体
}

注意事项：
对于类、方法来说，abstract关键字和final关键字不能同时使用，因为矛盾。

private 方法隐式地被指定为 final，如果在子类中定义的方法和基类中的一个 private 方法签名相同，此时子类的方法不是重写基类方法，而是在子类中定义了一个新的方法。

**3. 类**  

当final关键字用来修饰一个类的时候，格式：
public final class 类名称 {
    // ...
}

含义：当前这个类不能有任何的子类（不允许被继承）。（太监类）
注意：一个类如果是final的，那么其中所有的成员方法都无法进行覆盖重写（因为没儿子。）

## static

如果一个成员变量使用了static关键字，那么这个变量不再属于对象自己，而是属于所在的类。多个对象共享同一份数据（省内存）。静态代码在类的初始化阶段被初始化，非静态代码则在类的使用阶段(也就是实例化一个类的时候)才会被初始化。

**1. 静态变量**  

- 静态变量：又称为类变量，也就是说这个变量属于类的，类所有的实例都共享静态变量，可以直接通过类名来访问它。静态变量在内存中只存在一份。
- 实例变量：每创建一个实例就会产生一个实例变量，它与该实例同生共死。

```java
public class A {

    private int x;         // 实例变量
    private static int y;  // 静态变量

    public static void main(String[] args) {
        // int x = A.x;  // Non-static field 'x' cannot be referenced from a static context
        A a = new A();
        int x = a.x;
        int y = A.y;
    }
}
```

例子：

<div align="center"><img width="320px" src="../../pics/11c6229b-21e9-4e52-8116-c4148e95d1ed.png" width="500px"></img></div>

**2. 静态方法**  

一旦使用static修饰成员方法，那么这就成为了静态方法。静态方法不属于对象，而是属于类的。静态方法在类加载的时候就存在了，它不依赖于任何实例。所以静态方法必须有实现，也就是说它不能是抽象方法。

```java
public abstract class A {
    public static void func1(){
    }
    // public abstract static void func2();  // Illegal combination of modifiers: 'abstract' and 'static'
}
```

只能访问所属类的静态字段和静态方法，方法中不能有 this （对于本类中的静态方法，可省略类名字而直接调用）和 super 关键字，因此这两个关键字与具体对象关联。

```java
public class A {

    private static int x;
    private int y;

    public static void func1(){
        int a = x;
        // int b = y;  // Non-static field 'y' cannot be referenced from a static context
        // int b = this.y;     // 'A.this' cannot be referenced from a static context
    }
}
```



如果没有static关键字，那么必须首先创建对象，然后通过对象才能使用它。

如果有了static关键字，那么不需要创建对象，直接就能通过类名称来使用它。

无论是成员变量，还是成员方法。如果有了static，都推荐使用类名称进行调用。

静态变量：类名称.静态变量

静态方法：类名称.静态方法()

注意事项：
1. 静态不能直接访问非静态。
原因：因为在内存当中是【先】有的静态内容，【后】有的非静态内容。
“先人不知道后人，但是后人知道先人。”
2. 静态方法当中不能用this。
原因：this代表当前对象，通过谁调用的方法，谁就是当前对象。

内存示意图：

<div align="center"><img width="320px" src="../../pics/8834a48f-1175-4098-8fd3-626a114e0b3b.png" width="500px"></img></div>

**3. 静态代码块**  

特点：当第一次用到本类时，静态代码块执行唯一的一次。静态内容总是优先于非静态，所以静态代码块比构造方法先执行。

静态代码块的典型用途：用来一次性地对静态成员变量进行赋值。

```java
public class A {
    static {
        System.out.println("123");
    }

    public static void main(String[] args) {
        A a1 = new A();
        A a2 = new A();
    }
}
```

```html
123
```

**4. 静态内部类**  

非静态内部类依赖于外部类的实例，也就是说需要先创建外部类实例，才能用这个实例去创建非静态内部类，而静态内部类不需要。

1. 静态内部类是由static修饰的内部类（**普通的类无法用static关键字修饰**）
2. 静态内部类也是类，在继承和实现接口方面，和普通的类都是一样的
3. 外部类可以访问静态内部类的private属性。静态内部类，不能访问外部类的非静态的方法和属性，（如果在静态类的方法里，实例化了外部类，通过引用，静态内部类可以访问外部类的private属性）
4. **静态内部类可以被实例化，外部类每次实例化都会创建一个新的静态内部类对象**，不管外部类的状态(是否创建)，可以直接使用它的内部类
5. 静态内部类不会常驻内存，静态变量和方法才会常驻内存的方法区

```java
public class OuterClass {
		// 非静态内部类
    class InnerClass {
    }
		
   // 定义静态内部类  有static关键字修饰隶属于类层级
    static class StaticInnerClass {
    }

    public static void main(String[] args) {
        // InnerClass innerClass = new InnerClass(); // 'OuterClass.this' cannot be referenced from a static context
        OuterClass outerClass = new OuterClass();
        InnerClass innerClass = outerClass.new InnerClass();
        StaticInnerClass staticInnerClass = new StaticInnerClass();
    }
}
```

静态内部类的使用方式：

　　静态内部类不能直接访问外部类的变量和方法。

　　静态内部类可以直接创建对象。

　　如果静态内部类访问外部类中与本类（里面存在的）同名成员变量或方法时，需要使用类名.的方式访问。例子如下：

```java
/**
 * 实现静态内部类的定义和使用
 */
public class StaticOuter {
    private static int cnt = 1;
    private int snt = 2;

    /**
     * 定义静态内部类  有static关键字修饰隶属于类层级
     */
    public static class StaticInner {
        private int ia = 3;
        private static int cnt = 10;
        public StaticInner(){
            System.out.println("静态内部类的构造方法啊！");
        }
        public void show(){
            System.out.println("ia = " + ia);
            System.out.println("cnt = " + cnt);
            StaticOuter so = new StaticOuter();
            System.out.println("snt = " + so.snt);
        }
        // 静态内部类访问外部类中与本类（里面存在的）同名成员变量或方法时，需要使用类名.的方式访问
        public void show2(int cnt){
            System.out.println("形参变量是：" + cnt);
            System.out.println("静态内部类变量是：" + StaticInner.cnt);
            System.out.println("外部类静态变量是：" + StaticOuter.cnt);
        }
        // 可以使用在静态内部类使用show3的方法访问外部普通成员变量
        public void show3(){
            System.out.println(new  StaticOuter().snt);
        }
    }
}
```

注意：

1. 静态内部类以及静态内部类里面的静态元素，不会随项目启动而被加载。
2. 静态内部类不会依赖于外部类的初始化而初始化(静态内部类初始化与外部类无关，外部类初始化，静态内部类不会初始化。 静态内部类初始化了，外部类也不会初始化，两者没联系)
3. 当访问静态内部类里面的静态元素的时候，静态内部类以及里面的静态元素才会被初始化

**5. 静态导包**  

在使用静态变量和方法时不用再指明 ClassName，从而简化代码，但可读性大大降低。

```java
import static com.xxx.ClassName.*
```

**6. 初始化顺序**  

静态变量和静态语句块优先于实例变量和普通语句块，静态变量和静态语句块的初始化顺序取决于它们在代码中的顺序。

```java
public static String staticField = "静态变量";
```

```java
static {
    System.out.println("静态语句块");
}
```

```java
public String field = "实例变量";
```

```java
{
    System.out.println("普通语句块");
}
```

最后才是构造函数的初始化。

```java
public InitialOrderTest() {
    System.out.println("构造函数");
}
```

存在继承的情况下，初始化顺序为：

- 父类（静态变量、静态语句块）
- 子类（静态变量、静态语句块）
- 父类（实例变量、普通语句块）
- 父类（构造函数）
- 子类（实例变量、普通语句块）
- 子类（构造函数）

# 五、Object 通用方法

## 构造方法

构造方法是专门用来创建对象的方法，当用关键字new创建对象时，其实就是调用构造方法。

格式：

```
修饰符 类名称（参数类型 参数名称） {
	方法体
}
```

注意事项：

- 构造方法的名称必须和所在的类名称完全一样
- 构造方法不要写返回值类型，也没有return
- 如果没有写构造方法，编译器会使用默认的构造方法，没有参数/方法
- 一旦至少写了一个构造方法，编译器则不会使用默认方法
- 构造方法也是可以进行重载的（方法名称相同，参数列表不同）

```
/*
一个标准的类通常要拥有下面四个组成部分：

1. 所有的成员变量都要使用private关键字修饰
2. 为每一个成员变量编写一对儿Getter/Setter方法
3. 编写一个无参数的构造方法
4. 编写一个全参数的构造方法

这样标准的类也叫做Java Bean
 */
```

## 匿名对象

创建对象的标准格式：类名称 对象名 = new 类名称();

匿名对象就是只有右边的对象，没有左边的名字和赋值运算符。

new 类名称();

注意事项：匿名对象只能使用唯一的一次，下次再用不得不再创建一个新对象。

使用建议：如果确定有一个对象只需要使用唯一的一次，就可以用匿名对象。

虽然是创建对象的简化写法，但是应用场景非常有限。

```java
public class Person {

    String name;

    public void showName() {
        System.out.println("我叫：" + name);
    }

}

public class Demo01Anonymous {

    public static void main(String[] args) {
        // 左边的one就是对象的名字
        Person one = new Person();
        one.name = "高圆圆";
        one.showName(); // 我叫高圆圆
        System.out.println("===============");

        // 匿名对象
        new Person().name = "赵又廷";
        new Person().showName(); // 我叫：null
    }

}

public class Demo02Anonymous {

    public static void main(String[] args) {
        // 普通使用方式
//        Scanner sc = new Scanner(System.in);
//        int num = sc.nextInt();

        // 匿名对象的方式
//        int num = new Scanner(System.in).nextInt();
//        System.out.println("输入的是：" + num);

        // 使用一般写法传入参数
//        Scanner sc = new Scanner(System.in);
//        methodParam(sc);

        // 使用匿名对象来进行传参
//        methodParam(new Scanner(System.in));

        Scanner sc = methodReturn();
        int num = sc.nextInt();
        System.out.println("输入的是：" + num);
    }

    public static void methodParam(Scanner sc) {
        int num = sc.nextInt();
        System.out.println("输入的是：" + num);
    }

    public static Scanner methodReturn() {
//        Scanner sc = new Scanner(System.in);
//        return sc;
        return new Scanner(System.in);
    }

}
```

## 概览

`java.lang.Object`类是Java语言中的根类，即所有类的父类。它中描述的所有方法子类都可以使用。在对象实例化的时候，最终找的父类就是Object。

如果一个类没有特别指定父类，	那么默认则继承自Object类。例如：

```java
public class MyClass /*extends Object*/ {
  	// ...
}
```

根据JDK源代码及Object类的API文档，Object类当中包含的方法有11个。

```java
public native int hashCode()

public boolean equals(Object obj)

protected native Object clone() throws CloneNotSupportedException

public String toString()

public final native Class<?> getClass()

protected void finalize() throws Throwable {}

public final native void notify()

public final native void notifyAll()

public final native void wait(long timeout) throws InterruptedException

public final void wait(long timeout, int nanos) throws InterruptedException

public final void wait() throws InterruptedException
```

## equals()

 **== 的作用：**
　　基本类型：比较的就是值是否相同
　　引用类型：比较的就是地址值是否相同
**equals 的作用:**
　　引用类型：默认情况下，比较的是地址值。

String 中 **==** 比较引用地址是否相同，**equals()** 比较字符串的内容是否相同：

**方法摘要**

* `public boolean equals(Object obj)`：指示其他某个对象是否与此对象“相等”。

调用成员方法equals并指定参数为另一个对象，则可以判断这两个对象是否是相同的。这里的“相同”有默认和自定义两种方式。

**默认地址比较**

如果没有覆盖重写equals方法，那么Object类中默认进行`==`运算符的对象地址比较，只要不是同一个对象，结果必然为false。

**对象内容比较**

如果希望进行对象的内容比较，即所有或指定的部分成员变量相同就判定两个对象相同，则可以覆盖重写equals方法。例如：

```java
import java.util.Objects;

public class Person {	
	private String name;
	private int age;
	
    @Override
    public boolean equals(Object o) {
        // 如果对象地址一样，则认为相同
        if (this == o)
            return true;
        // 如果参数为空，或者类型信息不一样，则认为不同
        if (o == null || getClass() != o.getClass())
            return false;
        // 转换为当前类型
        Person person = (Person) o;
        // 要求基本类型相等，并且将引用类型交给java.util.Objects类的equals静态方法取用结果
        return age == person.age && Objects.equals(name, person.name);
    }
}
```

这段代码充分考虑了对象为空、类型一致等问题，但方法内容并不唯一。大多数IDE都可以自动生成equals方法的代码内容。

Person类默认继承了Object类,所以可以使用Object类的equals方法

boolean equals(Object obj) 指示其他某个对象是否与此对象“相等”。

equals方法源码:

​    public boolean equals(Object obj) {

​        return (this == obj);

​    }

​    参数:

​        Object obj:可以传递任意的对象

​        == 比较运算符,返回的是一个布尔值 true false

​        基本数据类型:比较的是值

​        引用数据类型:比较的是两个对象的地址值

   this是谁?那个对象调用的方法,方法中的this就是那个对象;p1调用的equals方法所以this就是p1

   obj是谁?传递过来的参数p2

   this==obj -->p1==p2



Object类的equals方法,默认比较的是两个对象的地址值,没有意义

所以我们要重写equals方法,比较两个对象的属性(name,age)

问题:

​    隐含着一个多态

​    多态的弊端:无法使用子类特有的内容(属性和方法)

​    Object obj = p2 = new Person("古力娜扎",19);

​    解决:可以使用向下转型(强转)把obj类型转换为Person

**1. 等价关系**  

两个对象具有等价关系，需要满足以下五个条件：

Ⅰ 自反性

```java
x.equals(x); // true
```

Ⅱ 对称性

```java
x.equals(y) == y.equals(x); // true
```

Ⅲ 传递性

```java
if (x.equals(y) && y.equals(z))
    x.equals(z); // true;
```

Ⅳ 一致性

多次调用 equals() 方法结果不变

```java
x.equals(y) == x.equals(y); // true
```

Ⅴ 与 null 的比较

对任何不是 null 的对象 x 调用 x.equals(null) 结果都为 false

```java
x.equals(null); // false;
```

**2. 等价与相等**  

- 对于基本类型，== 判断两个值是否相等，基本类型没有 equals() 方法。
- 对于引用类型，== 判断两个变量是否引用同一个对象，而 equals() 判断引用的对象是否等价。

```java
Integer x = new Integer(1);
Integer y = new Integer(1);
System.out.println(x.equals(y)); // true
System.out.println(x == y);      // false
```

**3. 实现**  

- 检查是否为同一个对象的引用，如果是直接返回 true；
- 检查是否是同一个类型，如果不是，直接返回 false；
- 将 Object 对象进行转型；
- 判断每个关键域是否相等。

```java
public class EqualExample {

    private int x;
    private int y;
    private int z;

    public EqualExample(int x, int y, int z) {
        this.x = x;
        this.y = y;
        this.z = z;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;

        EqualExample that = (EqualExample) o;

        if (x != that.x) return false;
        if (y != that.y) return false;
        return z == that.z;
    }
}
```

## hashCode()

```java
public class Person extends  Object{
    //重写hashCode方法

    @Override
    public int hashCode() {
        return  1;
    }
}

/*
    哈希值:是一个十进制的整数,由系统随机给出(就是对象的地址值,是一个逻辑地址,是模拟出来得到地址,不是数据实际存储的物理地址)
    在Object类有一个方法,可以获取对象的哈希值
    int hashCode() 返回该对象的哈希码值。
    hashCode方法的源码:
        public native int hashCode();
        native:代表该方法调用的是本地操作系统的方法
 */
public class Demo01HashCode {
    public static void main(String[] args) {
        //Person类继承了Object类,所以可以使用Object类的hashCode方法
        Person p1 = new Person();
        int h1 = p1.hashCode();
        System.out.println(h1);//1967205423  | 1

        Person p2 = new Person();
        int h2 = p2.hashCode();
        System.out.println(h2);//42121758   |  1

        /*
            toString方法的源码:
                return getClass().getName() + "@" + Integer.toHexString(hashCode());
         */
        System.out.println(p1);//com.itheima.demo03.hashCode.Person@75412c2f
        System.out.println(p2);//com.itheima.demo03.hashCode.Person@282ba1e
        System.out.println(p1==p2);//false

        /*
            String类的哈希值
                String类重写Obejct类的hashCode方法
         */
        String s1 = new String("abc");
        String s2 = new String("abc");
        System.out.println(s1.hashCode());//96354
        System.out.println(s2.hashCode());//96354

        System.out.println("重地".hashCode());//1179395
        System.out.println("通话".hashCode());//1179395
    }
}
```

hashCode() 返回哈希值，而 equals() 是用来判断两个对象是否等价。等价的两个对象散列值一定相同，但是散列值相同的两个对象不一定等价，这是因为计算哈希值具有随机性，两个值不同的对象可能计算出相同的哈希值。

在覆盖 equals() 方法时应当总是覆盖 hashCode() 方法，保证等价的两个对象哈希值也相等。

HashSet  和 HashMap 等集合类使用了 hashCode()  方法来计算对象应该存储的位置，因此要将对象添加到这些集合类中，需要让对应的类实现 hashCode()  方法。

下面的代码中，新建了两个等价的对象，并将它们添加到 HashSet 中。我们希望将这两个对象当成一样的，只在集合中添加一个对象。但是 EqualExample 没有实现 hashCode() 方法，因此这两个对象的哈希值是不同的，最终导致集合添加了两个等价的对象。

```java
EqualExample e1 = new EqualExample(1, 1, 1);
EqualExample e2 = new EqualExample(1, 1, 1);
System.out.println(e1.equals(e2)); // true
HashSet<EqualExample> set = new HashSet<>();
set.add(e1);
set.add(e2);
System.out.println(set.size());   // 2
```

理想的哈希函数应当具有均匀性，即不相等的对象应当均匀分布到所有可能的哈希值上。这就要求了哈希函数要把所有域的值都考虑进来。可以将每个域都当成 R 进制的某一位，然后组成一个 R 进制的整数。

R 一般取 31，因为它是一个奇素数，如果是偶数的话，当出现乘法溢出，信息就会丢失，因为与 2 相乘相当于向左移一位，最左边的位丢失。并且一个数与 31 相乘可以转换成移位和减法：`31*x == (x<<5)-x`，编译器会自动进行这个优化。

```java
@Override
public int hashCode() {
    int result = 17;
    result = 31 * result + x;
    result = 31 * result + y;
    result = 31 * result + z;
    return result;
}
```

## toString()

**方法摘要**

* `public String toString()`：返回该对象的字符串表示。

toString方法返回该对象的字符串表示，其实该字符串内容就是对象的类型+@+内存地址值。

默认返回 ToStringExample@4554617c 这种形式，其中 @ 后面的数值为散列码的无符号十六进制表示。

```java
public class ToStringExample {

    private int number;

    public ToStringExample(int number) {
        this.number = number;
    }
}
```

```java
ToStringExample example = new ToStringExample(123);
System.out.println(example.toString());
```

```html
ToStringExample@4554617c
```

直接打印对象的地址值没有意义,需要重写Object类中的toString方法

看一个类是否重写了toString,直接打印这个类的对象即可,如果没有重写toString方法那么打印的是对象的地址值。

由于toString方法返回的结果是内存地址，而在开发中，经常需要按照对象的属性得到相应的字符串表现形式，因此也需要重写它。

**覆盖重写**

如果不希望使用toString方法的默认行为，则可以对它进行覆盖重写。例如自定义的Person类：

```java
public class Person {  
    private String name;
    private int age;

    @Override
    public String toString() {
        return "Person{" + "name='" + name + '\'' + ", age=" + age + '}';
    }

    // 省略构造器与Getter Setter
}
```

在IntelliJ IDEA中，可以点击`Code`菜单中的`Generate...`，也可以使用快捷键`alt+insert`，点击`toString()`选项。选择需要包含的成员变量并确定。

## clone()

**1. cloneable**  

clone() 是 Object 的 protected 方法，它不是 public，一个类不显式去重写 clone()，其它类就不能直接去调用该类实例的 clone() 方法。

```java
public class CloneExample {
    private int a;
    private int b;
}
```

```java
CloneExample e1 = new CloneExample();
// CloneExample e2 = e1.clone(); // 'clone()' has protected access in 'java.lang.Object'
```

重写 clone() 得到以下实现：

```java
public class CloneExample {
    private int a;
    private int b;

    @Override
    public CloneExample clone() throws CloneNotSupportedException {
        return (CloneExample)super.clone();
    }
}
```

```java
CloneExample e1 = new CloneExample();
try {
    CloneExample e2 = e1.clone();
} catch (CloneNotSupportedException e) {
    e.printStackTrace();
}
```

```html
java.lang.CloneNotSupportedException: CloneExample
```

以上抛出了 CloneNotSupportedException，这是因为 CloneExample 没有实现 Cloneable 接口。

应该注意的是，clone() 方法并不是 Cloneable 接口的方法，而是 Object 的一个 protected 方法。Cloneable 接口只是规定，如果一个类没有实现 Cloneable 接口又调用了 clone() 方法，就会抛出 CloneNotSupportedException。

```java
public class CloneExample implements Cloneable {
    private int a;
    private int b;

    @Override
    public Object clone() throws CloneNotSupportedException {
        return super.clone();
    }
}
```

**2. 浅拷贝**  

拷贝对象和原始对象的引用类型引用同一个对象。

```java
public class ShallowCloneExample implements Cloneable {

    private int[] arr;

    public ShallowCloneExample() {
        arr = new int[10];
        for (int i = 0; i < arr.length; i++) {
            arr[i] = i;
        }
    }

    public void set(int index, int value) {
        arr[index] = value;
    }

    public int get(int index) {
        return arr[index];
    }

    @Override
    protected ShallowCloneExample clone() throws CloneNotSupportedException {
        return (ShallowCloneExample) super.clone();
    }
}
```

```java
ShallowCloneExample e1 = new ShallowCloneExample();
ShallowCloneExample e2 = null;
try {
    e2 = e1.clone();
} catch (CloneNotSupportedException e) {
    e.printStackTrace();
}
e1.set(2, 222);
System.out.println(e2.get(2)); // 222
```

**3. 深拷贝**  

拷贝对象和原始对象的引用类型引用不同对象。

```java
public class DeepCloneExample implements Cloneable {

    private int[] arr;

    public DeepCloneExample() {
        arr = new int[10];
        for (int i = 0; i < arr.length; i++) {
            arr[i] = i;
        }
    }

    public void set(int index, int value) {
        arr[index] = value;
    }

    public int get(int index) {
        return arr[index];
    }

    @Override
    protected DeepCloneExample clone() throws CloneNotSupportedException {
        DeepCloneExample result = (DeepCloneExample) super.clone();
        result.arr = new int[arr.length];
        for (int i = 0; i < arr.length; i++) {
            result.arr[i] = arr[i];
        }
        return result;
    }
}
```

```java
DeepCloneExample e1 = new DeepCloneExample();
DeepCloneExample e2 = null;
try {
    e2 = e1.clone();
} catch (CloneNotSupportedException e) {
    e.printStackTrace();
}
e1.set(2, 222);
System.out.println(e2.get(2)); // 2
```

**4. clone() 的替代方案**  

使用 clone() 方法来拷贝一个对象即复杂又有风险，它会抛出异常，并且还需要类型转换。Effective Java 书上讲到，最好不要去使用 clone()，可以使用拷贝构造函数或者拷贝工厂来拷贝一个对象。

```java
public class CloneConstructorExample {

    private int[] arr;

    public CloneConstructorExample() {
        arr = new int[10];
        for (int i = 0; i < arr.length; i++) {
            arr[i] = i;
        }
    }

    public CloneConstructorExample(CloneConstructorExample original) {
        arr = new int[original.arr.length];
        for (int i = 0; i < original.arr.length; i++) {
            arr[i] = original.arr[i];
        }
    }

    public void set(int index, int value) {
        arr[index] = value;
    }

    public int get(int index) {
        return arr[index];
    }
}
```

```java
CloneConstructorExample e1 = new CloneConstructorExample();
CloneConstructorExample e2 = new CloneConstructorExample(e1);
e1.set(2, 222);
System.out.println(e2.get(2)); // 2
```

## Objects类

在刚才IDEA自动重写equals代码中，使用到了`java.util.Objects`类，那么这个类是什么呢？

在**JDK7**添加了一个Objects工具类，它提供了一些方法来操作对象，它由一些静态的实用方法组成，这些方法是null-save（空指针安全的）或null-tolerant（容忍空指针的），用于计算对象的hashcode、返回对象的字符串表示形式、比较两个对象。

在比较两个对象的时候，Object的equals方法容易抛出空指针异常，而Objects类中的equals方法就优化了这个问题。方法如下：

* `public static boolean equals(Object a, Object b)`:判断两个对象是否相等。

我们可以查看一下源码，学习一下：

~~~java
public static boolean equals(Object a, Object b) {  
    return (a == b) || (a != null && a.equals(b));  
}
~~~



# 六、面向对象三大特征-封装

方法和关键字均是封装特性的体现：将细节信息隐藏，对外界不可见，只管调用。

Java中有四种权限修饰符：

|                        | public | protected | (default) | private |
| ---------------------- | ------ | --------- | --------- | ------- |
| 同一个类（我自己）     | YES    | YES       | YES       | YES     |
| 同一个包（我邻居）     | YES    | YES       | YES       | NO      |
| 不同包子类（我儿子）   | YES    | YES       | NO        | NO      |
| 不同包非子类（陌生人） | YES    | NO        | NO        | NO      |

注意事项：(default)并不是关键字“default”，而是根本不写。

## 访问权限

可以对类或类中的成员（字段和方法）加上访问修饰符。

- 类可见表示其它类可以用这个类创建实例对象。
- 成员可见表示其它类可以用这个类的实例对象访问到该成员；

protected 用于修饰成员，表示在继承体系中成员对于子类可见，但是这个访问修饰符对于类没有意义。

设计良好的模块会隐藏所有的实现细节，把它的 API 与它的实现清晰地隔离开来。模块之间只通过它们的 API 进行通信，一个模块不需要知道其他模块的内部工作情况，这个概念被称为信息隐藏或封装。因此访问权限应当尽可能地使每个类或者成员不被外界访问。

如果子类（派生类）的方法重写了父类（基类/超类）的方法，那么子类中该方法的访问级别不允许低于父类的访问级别。这是为了确保可以使用父类实例的地方都可以使用子类实例去代替，也就是确保满足里氏替换原则。

字段决不能是公有的，因为这么做的话就失去了对这个字段修改行为的控制，客户端可以对其随意修改。例如下面的例子中，AccessExample 拥有 id 公有字段，如果在某个时刻，我们想要使用 int 存储 id 字段，那么就需要修改所有的客户端代码。

```java
public class AccessExample {
    public String id;
}
```

可以使用公有的 getter 和 setter 方法来替换公有字段，这样的话就可以控制对字段的修改行为。

```java
public class AccessExample {

    private int id;

    public String getId() {
        return id + "";
    }

    public void setId(String id) {
        this.id = Integer.valueOf(id);
    }
}
```

但是也有例外，如果是包级私有的类或者私有的嵌套类，那么直接暴露成员不会有特别大的影响。

```java
public class AccessWithInnerClassExample {

    private class InnerClass {
        int x;
    }

    private InnerClass innerClass;

    public AccessWithInnerClassExample() {
        innerClass = new InnerClass();
    }

    public int getValue() {
        return innerClass.x;  // 直接访问
    }
}
```

# 七、面向对象三大特征-继承

## 变量与方法

在继承的关系中，“子类就是一个父类”。也就是说，子类可以被当做父类看待。

例如父类是员工，子类是讲师，那么“讲师就是一个员工”。关系：is-a。

定义父类的格式：（一个普通的类定义）

public class 父类名称 {

​    // ...
}

定义子类的格式：

public class 子类名称 extends 父类名称 {

​    // ...
}

注意事项：

- 在父子类的继承关系当中，如果成员变量重名，则创建子类对象时，访问有两种方式：

  直接通过子类对象访问成员变量： 等号左边是谁，就优先用谁，没有则向上找。

  间接通过成员方法访问成员变量： 该方法属于谁，就优先用谁，没有则向上找。

  ```java
  public class Demo01ExtendsField {
  
      public static void main(String[] args) {
          Fu fu = new Fu(); // 创建父类对象
          System.out.println(fu.numFu); // 只能使用父类的东西，没有任何子类内容
          System.out.println("===========");
  
          Zi zi = new Zi();
  
          System.out.println(zi.numFu); // 10
          System.out.println(zi.numZi); // 20
          System.out.println("===========");
  
          // 等号左边是谁，就优先用谁
          System.out.println(zi.num); // 优先子类，200
  //        System.out.println(zi.abc); // 到处都没有，编译报错！
          System.out.println("===========");
  
          // 这个方法是子类的，优先用子类的，没有再向上找
          zi.methodZi(); // 200
          // 这个方法是在父类当中定义的，
          zi.methodFu(); // 100
      }
  
  }
  ```

- 变量重名

  局部变量：         直接写成员变量名
  本类的成员变量：    this.成员变量名
  父类的成员变量：    super.成员变量名

  ```java
  public class Fu {
  
      int num = 10;
  
  }
  ```

  ```java
  public class Zi extends Fu {
  
      int num = 20;
  
      public void method() {
          int num = 30;
          System.out.println(num); // 30，局部变量
          System.out.println(this.num); // 20，本类的成员变量
          System.out.println(super.num); // 10，父类的成员变量
      }
  
  }
  ```

- 方法重名

  在父子类的继承关系当中，创建子类对象，访问成员方法的规则：创建的对象是谁，就优先用谁，如果没有则向上找。

  **注意事项**：

  - 无论是成员方法还是成员变量，如果没有都是向上找父类，绝对不会向下找子类的。

  - 必须保证父子类之间方法的名称相同，参数列表也相同。

  - @Override：写在方法前面，用来检测是不是有效的正确覆盖重写。这个注解就算不写，只要满足要求，也是正确的方法覆盖重写。

  - 子类方法的返回值必须【小于等于】父类方法的返回值范围。

    小扩展提示：java.lang.Object类是所有类的公共最高父类（祖宗类），java.lang.String就是Object的子类。

  - 子类方法的权限必须【大于等于】父类方法的权限修饰符。

    小扩展提示：public > protected > (default) > private ((default)不是关键字default，而是什么都不写，留空)

  重写（Override）概念：在继承关系当中，方法的名称一样，参数列表也一样。

  重写（Override）：方法的名称一样，参数列表【也一样】。覆盖、覆写。

  重载（Overload）：方法的名称一样，参数列表【不一样】。

  方法的覆盖重写特点：创建的是子类对象，则优先用子类方法。

  ```
  方法覆盖重写的注意事项：
  
  1. 必须保证父子类之间方法的名称相同，参数列表也相同。
  @Override：写在方法前面，用来检测是不是有效的正确覆盖重写。
  这个注解就算不写，只要满足要求，也是正确的方法覆盖重写。
  
  2. 子类方法的返回值必须【小于等于】父类方法的返回值范围。
  小扩展提示：java.lang.Object类是所有类的公共最高父类（祖宗类），java.lang.String就是Object的子类。
  
  3. 子类方法的权限必须【大于等于】父类方法的权限修饰符。
  小扩展提示：public > protected > (default) > private
  备注：(default)不是关键字default，而是什么都不写，留空。
  ```

  ```java
  // 定义一个新手机，使用老手机作为父类
  public class NewPhone extends Phone {
  
      @Override
      public void show() {
          super.show(); // 把父类的show方法拿过来重复利用
          // 自己子类再来添加更多内容
          System.out.println("显示姓名");
          System.out.println("显示头像");
      }
  }
  ```


## super

- super关键字的用法有三种：
  1. 在子类的成员方法中，访问父类的成员变量。
  2. 在子类的成员方法中，访问父类的成员方法。
  3. 在子类的构造方法中，访问父类的构造方法。

- 继承关系中，父子类构造方法的访问特点：

  1. 可以使用 super() 函数访问父类的构造函数，从而委托父类完成一些初始化的工作。应该注意到，**子类一定会调用父类的构造函数来完成初始化工作，一般是调用父类的默认构造函数(隐含的“super()”调用)**。

     如果子类需要调用父类其它构造函数，那么就可以使用 super() 函数。所以一定是先调用的父类构造，后执行的子类构造。

  2. 子类构造可以通过super关键字来调用父类重载构造。

  3. super的父类构造调用，必须是子类构造方法的第一个语句。不能一个子类构造调用多次super构造。

  总结：

  子类必须调用父类构造方法，不写则赠送super()；写了则用写的指定的super调用，super只能有一个，还必须是第一个。

- 访问父类的成员：如果子类重写了父类的某个方法，可以通过使用 super 关键字来引用父类的方法实现。

```java
public class SuperExample {

    protected int x;
    protected int y;

    public SuperExample(int x, int y) {
        this.x = x;
        this.y = y;
    }

    public void func() {
        System.out.println("SuperExample.func()");
    }
}
```

```java
public class SuperExtendExample extends SuperExample {

    private int z;

    public SuperExtendExample(int x, int y, int z) {
        super(x, y);
        this.z = z;
    }

    @Override
    public void func() {
        super.func();
        System.out.println("SuperExtendExample.func()");
    }
}
```

```java
SuperExample e = new SuperExtendExample(1, 2, 3);
e.func();
```

```html
SuperExample.func()
SuperExtendExample.func()
```

[Using the Keyword super](https://docs.oracle.com/javase/tutorial/java/IandI/super.html)

## this

- 当方法的局部变量和类的成员变量重名的时候，根据“就近原则”，优先使用局部变量。

- 如果需要访问本类中的成员变量，需要使用：this.成员变量名。

- 通过谁调用的方法，谁就是this（本质上就是对象，直接打印this是对象的地址）

- super关键字用来访问父类内容，而this关键字用来访问本类内容。用法也有三种：

  1. 在本类的成员方法中，访问本类的成员变量。
  2. 在本类的成员方法中，访问本类的另一个成员方法。
  3. 在本类的构造方法中，访问本类的另一个构造方法。
     在第三种用法当中要注意：
     A. this(...)调用也必须是构造方法的第一个语句，唯一一个。
     B. super和this两种构造调用，不能同时使用。

  ```
  public class Zi extends Fu {
  
      int num = 20;
  
      public Zi() {
  //        super(); // 这一行不再赠送
          this(123); // 本类的无参构造，调用本类的有参构造
  //        this(1, 2); // 错误写法！
      }
  
      public Zi(int n) {
          this(1, 2);
      }
  
      public Zi(int n, int m) {
      }
  
      public void showNum() {
          int num = 10;
          System.out.println(num); // 局部变量
          System.out.println(this.num); // 本类中的成员变量
          System.out.println(super.num); // 父类中的成员变量
      }
  
      public void methodA() {
          System.out.println("AAA");
      }
  
      public void methodB() {
          this.methodA();
          System.out.println("BBB");
      }
  }
  ```
  
  图解this与super
  
  <div align="center"> <img src="../../pics/8852d235-0c8e-4d8b-8124-8741633a6cf5.png" width="800"/> </div><br>

## 继承的特点

- Java语言是单继承的：一个类的直接父类只有唯一一个
- Java语言可以多级继承
- 一个类的直接父类是唯一的，但一个父类可以拥有很多个子类

<div align="center"> <img src="../../pics/c1e16f34-aa4e-4fd1-91ff-ee58abc04f83.png" width="800"/> </div><br>

# 八、抽象类与接口

## 抽象类

**1. 概述**

抽象方法：就是加上abstract关键字，然后去掉大括号，直接分号结束。
抽象类：抽象方法所在的类，必须是抽象类才行。在class之前写上abstract即可。

<div align="center"> <img src="../../pics/01ac6074-e1c3-4806-83f0-11965d882d0c.png" width="500"/> </div><br>

如何使用抽象类和抽象方法：
1. 不能直接创建new抽象类对象。

2. 必须用一个子类来继承抽象父类。

3. 子类必须覆盖重写抽象父类当中所有的抽象方法。
    覆盖重写（实现）：子类去掉抽象方法的abstract关键字，然后补上方法体大括号。

4. 创建子类对象进行使用。

   ```java
   public abstract class Animal {
   
       // 这是一个抽象方法，代表吃东西，但是具体吃什么（大括号的内容）不确定。
       public abstract void eat();
   
       // 这是普通的成员方法
   //    public void normalMethod() {
   //    }
   }
   ```

注意事项

1. 抽象类不能创建对象，创建编译无法通过而报错。只能创建其非抽象子类的对象。
2. 抽象类中，可以有构造方法，是供子类创建对象时初始化非类成员使用的。
3. 抽象类中，不一定包含抽象方法（适配器模式可用），但是有抽象方法的类必定是抽象类。
4. 抽象类的子类，必须重写抽象父类中所有的抽象方法。否则，编译无法通过而报错，除非该子类也是抽象类。

**2. 抽象类**

抽象类和抽象方法都使用 abstract 关键字进行声明。如果一个类中包含抽象方法，那么这个类必须声明为抽象类。

抽象类和普通类最大的区别是，抽象类不能被实例化，只能被继承。

```java
public abstract class AbstractClassExample {

    protected int x;
    private int y;

    public abstract void func1();

    public void func2() {
        System.out.println("func2");
    }
}
```

```java
public class AbstractExtendClassExample extends AbstractClassExample {
    @Override
    public void func1() {
        System.out.println("func1");
    }
}
```

```java
// AbstractClassExample ac1 = new AbstractClassExample(); // 'AbstractClassExample' is abstract; cannot be instantiated
AbstractClassExample ac2 = new AbstractExtendClassExample();
ac2.func1();
```



## 接口

**1. 概述**

- 接口是抽象类的延伸，接口就是多个类的公共规范。接口是一种引用数据类型，最重要的内容就是其中的：抽象方法。

- 接口的格式：

  ```java
  public interface 接口名称 {
      // 接口内容
  }
  ```

  备注：换成了关键字interface之后，编译生成的字节码文件仍然是：.java --> .class。

- 如果是Java 7，那么接口中可以包含的内容有：常量    抽象方法

- 如果是Java 8，还可以额外包含有：默认方法    静态方法

  在 Java 8 之前，它可以看成是一个完全抽象的类，也就是说它不能有任何的方法实现。

  从 Java 8 开始，接口也可以拥有默认的方法实现，这是因为不支持默认方法的接口的维护成本太高了。在 Java 8 之前，如果一个接口想要添加新的方法，那么要修改所有实现了该接口的类，让它们都实现新增的方法。

- 如果是Java 9，还可以额外包含有：私有方法

  接口的成员（字段 + 方法）默认都是 public 的，并且不允许定义为 private 或者 protected。从 Java 9 开始，允许将方法定义为 private，这样就能定义某些复用的代码又不会把方法暴露出去。

- 注意事项

  如果实现类并没有覆盖重写接口中所有的抽象方法，那么这个实现类自己就必须是抽象类。

  

- 在任何版本的Java中，接口都能定义抽象方法。格式：

  public abstract 返回值类型 方法名称(参数列表);

  注意事项：

  1. 接口当中的抽象方法，修饰符必须是两个固定的关键字：public abstract
  2. 这两个关键字修饰符，可以选择性地省略。
  3. 方法的三要素，可以随意定义。

  ```java
  public interface MyInterfaceAbstract {
  
      // 这是一个抽象方法
      public abstract void methodAbs1();
  
      // 这也是抽象方法
      abstract void methodAbs2();
  
      // 这也是抽象方法
      public void methodAbs3();
  
      // 这也是抽象方法
      void methodAbs4();
  }
  ```

  

- 接口使用步骤：

  1. 接口不能直接使用，必须有一个“实现类”来“实现”该接口。
     格式：

     ```java
     public class 实现类名称 implements 接口名称 {
      // ...
     }
     ```

  2. 接口的实现类必须覆盖重写（实现）接口中所有的抽象方法。
     实现：去掉abstract关键字，加上方法体大括号。

  3. 创建实现类的对象，进行使用。

  ```java
  public class MyInterfaceAbstractImpl implements MyInterfaceAbstract {
      @Override
      public void methodAbs1() {
          System.out.println("这是第一个方法！");
      }
  
      @Override
      public void methodAbs2() {
          System.out.println("这是第二个方法！");
      }
  
      @Override
      public void methodAbs3() {
          System.out.println("这是第三个方法！");
      }
  
      @Override
      public void methodAbs4() {
          System.out.println("这是第四个方法！");
      }
  }
  
  public class Demo01Interface {
  
      public static void main(String[] args) {
          // 错误写法！不能直接new接口对象使用。
  //        MyInterfaceAbstract inter = new MyInterfaceAbstract();
  
          // 创建实现类的对象使用
          MyInterfaceAbstractImpl impl = new MyInterfaceAbstractImpl();
          impl.methodAbs1();
          impl.methodAbs2();
      }
  
  }
  ```

  

- 从Java 8开始，接口里允许定义默认方法。格式：

  ```java
  public default 返回值类型 方法名称(参数列表) {
      方法体
  }
  ```

  备注：接口当中的默认方法，可以解决接口升级的问题。

  ```java
  public interface MyInterfaceDefault {
  
      // 抽象方法
      public abstract void methodAbs();
  
      // 新添加了一个抽象方法
  //    public abstract void methodAbs2();
  
      // 新添加的方法，改成默认方法
      public default void methodDefault() {
          System.out.println("这是新添加的默认方法");
      }
  
  }
  
  // 不受methodDefault方法的影响
  /*
  1. 接口的默认方法，可以通过接口实现类对象，直接调用。
  2. 接口的默认方法，也可以被接口实现类进行覆盖重写。
   */
  public class Demo02Interface {
  
      public static void main(String[] args) {
          // 创建了实现类对象
          MyInterfaceDefaultA a = new MyInterfaceDefaultA();
          a.methodAbs(); // 调用抽象方法，实际运行的是右侧实现类。
  
          // 调用默认方法，如果实现类当中没有，会向上找接口
          a.methodDefault(); // 这是新添加的默认方法
      }
  }
  ```



- 从Java 8开始，接口当中允许定义静态方法。格式：

  ```java
  public static 返回值类型 方法名称(参数列表) {
      方法体
  }
  ```

  提示：就是将abstract或者default换成static即可，带上方法体。

  ```java
  public interface MyInterfaceStatic {
  
      public static void methodStatic() {
          System.out.println("这是接口的静态方法！");
      }
  
  }
  
  public class MyInterfaceStaticImpl implements MyInterfaceStatic {
  }
  
  /*
  注意事项：不能通过接口实现类的对象来调用接口当中的静态方法。
  正确用法：通过接口名称，直接调用其中的静态方法。
  格式：
  接口名称.静态方法名(参数);
   */
  public class Demo03Interface {
  
      public static void main(String[] args) {
          // 创建了实现类对象
          MyInterfaceStaticImpl impl = new MyInterfaceStaticImpl();
  
          // 错误写法！
  //        impl.methodStatic();
  
          // 直接通过接口名称调用静态方法
          MyInterfaceStatic.methodStatic();
      }
  
  }
  ```

  

- 我们需要抽取一个共有方法，用来解决两个默认方法之间重复代码的问题。但是这个共有方法不应该让实现类使用，应该是私有化的。解决方案：
  从Java 9开始，接口当中允许定义私有方法。

  1. 普通私有方法，解决多个默认方法之间重复代码问题，格式：

    ```java
    private 返回值类型 方法名称(参数列表) {
     方法体
    }
    ```

  2. 静态私有方法，解决多个静态方法之间重复代码问题，格式：

    ```java
    private static 返回值类型 方法名称(参数列表) {
     方法体
    }
    ```

  ```java
  public interface MyInterfacePrivateA {
  
      public default void methodDefault1() {
          System.out.println("默认方法1");
          methodCommon();
      }
  
      public default void methodDefault2() {
          System.out.println("默认方法2");
          methodCommon();
      }
  
      private void methodCommon() {
          System.out.println("AAA");
          System.out.println("BBB");
          System.out.println("CCC");
      }
  
  }
  
  public interface MyInterfacePrivateB {
  
      public static void methodStatic1() {
          System.out.println("静态方法1");
          methodStaticCommon();
      }
  
      public static void methodStatic2() {
          System.out.println("静态方法2");
          methodStaticCommon();
      }
  
      private static void methodStaticCommon() {
          System.out.println("AAA");
          System.out.println("BBB");
          System.out.println("CCC");
      }
  
  }
  
  public class Demo04Interface {
  
      public static void main(String[] args) {
          MyInterfacePrivateB.methodStatic1();
          MyInterfacePrivateB.methodStatic2();
          // 错误写法！
  //        MyInterfacePrivateB.methodStaticCommon();
      }
  
  }
  ```



- 接口当中也可以定义“成员变量”，但是必须使用public static final三个关键字进行修饰。从效果上看，这其实就是接口的【常量】。格式public static final 数据类型 常量名称 = 数据值;
  备注：一旦使用final关键字进行修饰，说明不可改变。

  注意事项：
  1. 接口当中的常量，可以省略public static final，注意：不写也照样是这样。
  2. 接口当中的常量，必须进行赋值；不能不赋值。
  3. 接口中常量的名称，使用完全大写的字母，用下划线进行分隔。（推荐命名规则）

  ```java
  public interface MyInterfaceConst {
  
      // 这其实就是一个常量，一旦赋值，不可以修改
      public static final int NUM_OF_MY_CLASS = 12;
  
  }
  
  public class Demo05Interface {
  
      public static void main(String[] args) {
          // 访问接口当中的常量
          System.out.println(MyInterfaceConst.NUM_OF_MY_CLASS);
      }
  
  }
  ```

**2. 接口的注意事项**

1. 接口是没有静态代码块或者构造方法的。

2. 一个类的直接父类是唯一的（类与类之间是单继承的），但是一个类可以同时实现多个接口（类与接口之间是多实现的）。格式：

  ```java
  public class MyInterfaceImpl implements MyInterfaceA, MyInterfaceB {
   // 覆盖重写所有抽象方法
  }
  ```

3. 如果实现类所实现的多个接口当中，存在重复的抽象方法（例如MyInterfaceA和MyInterfaceB中有一样的抽象方法），那么只需要覆盖重写一次即可。

4. 如果实现类没有覆盖重写所有接口当中的所有抽象方法，那么实现类就必须是一个抽象类。

5. 如果实现类所实现的多个接口当中，存在重复的默认方法，那么实现类一定要对冲突的默认方法进行覆盖重写。

6. 一个类的直接父类当中的方法，和接口当中的默认方法产生了冲突，优先用父类当中的方法。（继承优先接口）例子：

   ```java
   public interface MyInterface {
   
       public default void method() {
           System.out.println("接口的默认方法");
       }
   
   }
   
   public class Fu {
   
       public void method() {
           System.out.println("父类方法");
       }
   
   }
   
   public class Zi extends Fu implements MyInterface {
   }
   
   public class Demo01Interface {
   
       public static void main(String[] args) {
           Zi zi = new Zi();
           zi.method();
       }
   
   }
   ```
   
7. 接口与接口之间是多继承的（类与类之间是单继承的。直接父类只有一个。类与接口之间是多实现的。一个类可以实现多个接口。）

   ```java
   public interface MyInterface extends MyInterfaceA, MyInterfaceB {
   
       public abstract void method();
   s
       @Override
       public default void methodDefault() {
   
       }
   }
   ```

   注意事项：

   1. 多个父接口当中的抽象方法如果重复，没关系。
   2. 多个父接口当中的默认方法如果重复，那么子接口必须进行默认方法的覆盖重写，【而且带着default关键字】。

接口与具体的实现类之间也存在多态性

```java
// 左边是接口名称，右边是实现类名称，这就是多态写法
List<String> list = new ArrayList<>();
```



**3. 比较**  

- 从设计层面上看，抽象类提供了一种 IS-A 关系，需要满足里式替换原则，即子类对象必须能够替换掉所有父类对象。而接口更像是一种 LIKE-A 关系，它只是提供一种方法实现规范，并不要求接口和实现接口的类具有 IS-A 关系。
- 从使用上来看，一个类可以实现多个接口，但是不能继承多个抽象类。
- 接口的字段只能是 static 和 final 类型的，而抽象类的字段没有这种限制。
- 接口的成员只能是 public 的，而抽象类的成员可以有多种访问权限。

**4. 使用选择**  

使用接口：

- 需要让不相关的类都实现一个方法，例如不相关的类都可以实现 Comparable 接口中的 compareTo() 方法；
- 需要使用多重继承。

使用抽象类：

- 需要在几个相关的类中共享代码。
- 需要能控制继承来的成员的访问权限，而不是都为 public。
- 需要继承非静态和非常量字段。

在很多情况下，接口优先于抽象类。因为接口没有抽象类严格的类层次结构要求，可以灵活地为一个类添加行为。并且从 Java 8 开始，接口也可以有默认的方法实现，使得修改接口的成本也变的很低。

- [Abstract Methods and Classes](https://docs.oracle.com/javase/tutorial/java/IandI/abstract.html)
- [深入理解 abstract class 和 interface](https://www.ibm.com/developerworks/cn/java/l-javainterface-abstract/)
- [When to Use Abstract Class and Interface](https://dzone.com/articles/when-to-use-abstract-class-and-intreface)
- [Java 9 Private Methods in Interfaces](https://www.journaldev.com/12850/java-9-private-methods-interfaces)

# 九、面向对象三大特征-多态

## 概述

继承或实现是多态的前提，如果没有继承或实现也就没有多态。类与类的继承/接口与接口的继承/接口与实现类。

用继承/实现举个例子：

<div align="center"> <img src="../../pics/148561c8-89bc-446b-8456-9d6b0da8d60b.png" width="500"/> </div><br>

**补充**

接口的出现是为了更好的实现多态，而多态的实现不一定需要依赖于接口。多态一般有三种，接口的多态，类的多态，方法的多态。方法的多态就类似于我们方法的重载。类的多态无非就是子类继承父类，并重写父类的方法，从而获得不同的实现。那么再来看接口，接口跟类基本是一样，实现接口并实现接口的方法。。不同的类实现接口可以有不同的方式从而表现不同的行为，就是接口的多态性啊。



代码当中体现多态性，其实就是一句话：父类引用指向子类对象（对于接口，则是左接口右实现类）。格式：

父类名称 对象名 = new 子类名称();

或者：

接口名称 对象名 = new 实现类名称();

```java
public class Fu {

    public void method() {
        System.out.println("父类方法");
    }

    public void methodFu() {
        System.out.println("父类特有方法");
    }

}

public class Zi extends Fu {

    @Override
    public void method() {
        System.out.println("子类方法");
    }
}

public class Demo01Multi {

    public static void main(String[] args) {
        // 使用多态的写法
        // 左侧父类的引用，指向了右侧子类的对象
        Fu obj = new Zi();
				// 调用抽象方法，实际运行的是右侧实现类。
        // 调用默认方法，如果实现类当中没有，会向上找接口
        obj.method();// 先去找子
        obj.methodFu();
    }
}
//////////////////////////////////////////////////
子类方法
父类特有方法
```

左“父”右“子”就是多态！！！！！！！

**9.1 多态中成员变量和方法的使用特点**

访问成员**变量**的两种方式:

1. 直接通过对象名称访问成员变量：看等号左边是谁，优先用谁，没有则向上找。
2. 间接通过成员方法访问成员变量：看该方法属于谁，优先用谁，没有则向上找

```java
public class Fu /*extends Object*/ {

    int num = 10;

    public void showNum() {
        System.out.println(num);
    }

    public void method() {
        System.out.println("父类方法");
    }

    public void methodFu() {
        System.out.println("父类特有方法");
    }

}

public class Zi extends Fu {

    int num = 20;

    int age = 16;

    @Override
    public void showNum() {
        System.out.println(num);
    }

    @Override
    public void method() {
        System.out.println("子类方法");
    }

    public void methodZi() {
        System.out.println("子类特有方法");
    }
}

public class Demo01MultiField {

    public static void main(String[] args) {
        // 使用多态的写法，父类引用指向子类对象
        Fu obj = new Zi();
        System.out.println(obj.num); // 父：10
//        System.out.println(obj.age); // 错误写法！
        System.out.println("=============");

        // 子类没有覆盖重写，就是父：10
        // 子类如果覆盖重写，就是子：20
        obj.showNum();
    }

}
```

在多态的代码当中，成员**方法**的访问规则是：看new的是谁，就优先用谁，没有则向上找。

口诀：编译看左边，运行看右边。

对比一下：
成员变量：编译看左边，运行还看左边。
成员方法：编译看左边，运行看右边。

```java
public class Demo02MultiMethod {

    public static void main(String[] args) {
        Fu obj = new Zi(); // 多态

        obj.method(); // 父子都有，优先用子
        obj.methodFu(); // 子类没有，父类有，向上找到父类

        // 编译看左边，左边是Fu，Fu当中没有methodZi方法，所以编译报错。
//        obj.methodZi(); // 错误写法！
    }

}
```

<div align="center"> <img src="../../pics/0b1940e7-e880-40d7-b396-a6dd4c79a52e.png" width="800"/> </div><br>

**9.3 对象的转型**

<div align="center"> <img src="../../pics/d8046c3b-3ae1-459a-87f2-acb053eeee22.png" width="800"/> </div><br>

```java
public abstract class Animal {

    public abstract void eat();

}

public class Cat extends Animal {
    @Override
    public void eat() {
        System.out.println("猫吃鱼");
    }

    // 子类特有方法
    public void catchMouse() {
        System.out.println("猫抓老鼠");
    }
}

public class Dog extends Animal {
    @Override
    public void eat() {
        System.out.println("狗吃SHIT");
    }

    public void watchHouse() {
        System.out.println("狗看家");
    }
}


/*
向上转型一定是安全的，没有问题的，正确的。但是也有一个弊端：
对象一旦向上转型为父类，那么就无法调用子类原本特有的内容。

解决方案：用对象的向下转型【还原】。
 */
public class Demo01Main {

    public static void main(String[] args) {
        // 对象的向上转型，就是：父类引用指向之类对象。
        Animal animal = new Cat(); // 本来创建的时候是一只猫
        animal.eat(); // 猫吃鱼

//        animal.catchMouse(); // 错误写法！

        // 向下转型，进行“还原”动作
        Cat cat = (Cat) animal;
        cat.catchMouse(); // 猫抓老鼠

        // 下面是错误的向下转型
        // 本来new的时候是一只猫，现在非要当做狗
        // 错误写法！编译不会报错，但是运行会出现异常：
        // java.lang.ClassCastException，类转换异常
        Dog dog = (Dog) animal;
    }

}
```

```java
/*
如何才能知道一个父类引用的对象，本来是什么子类？
格式：
对象 instanceof 类名称
这将会得到一个boolean值结果，也就是判断前面的对象能不能当做后面类型的实例。
 */
public class Demo02Instanceof {

    public static void main(String[] args) {
        Animal animal = new Dog(); // 本来是一只狗
        animal.eat(); // 狗吃SHIT

        // 如果希望掉用子类特有方法，需要向下转型
        // 判断一下父类引用animal本来是不是Dog
        if (animal instanceof Dog) {
            Dog dog = (Dog) animal;
            dog.watchHouse();
        }
        // 判断一下animal本来是不是Cat
        if (animal instanceof Cat) {
            Cat cat = (Cat) animal;
            cat.catchMouse();
        }

        giveMeAPet(new Dog());
    }

    public static void giveMeAPet(Animal animal) {
        if (animal instanceof Dog) {
            Dog dog = (Dog) animal;
            dog.watchHouse();
        }
        if (animal instanceof Cat) {
            Cat cat = (Cat) animal;
            cat.catchMouse();
        }
    }

}
```

## 重写与重载

**1. 重写（Override）**  

存在于继承体系中，指子类实现了一个与父类在方法声明上完全相同的一个方法。

为了满足里式替换原则，重写有以下三个限制：

- 子类方法的访问权限必须大于等于父类方法；
- 子类方法的返回类型必须是父类方法返回类型或为其子类型。
- 子类方法抛出的异常类型必须是父类抛出异常类型或为其子类型。

使用 @Override 注解，可以让编译器帮忙检查是否满足上面的三个限制条件。

下面的示例中，SubClass 为 SuperClass 的子类，SubClass 重写了 SuperClass 的 func() 方法。其中：

- 子类方法访问权限为 public，大于父类的 protected。
- 子类的返回类型为 ArrayList<Integer>，是父类返回类型 List<Integer> 的子类。
- 子类抛出的异常类型为 Exception，是父类抛出异常 Throwable 的子类。
- 子类重写方法使用 @Override 注解，从而让编译器自动检查是否满足限制条件。

```java
class SuperClass {
    protected List<Integer> func() throws Throwable {
        return new ArrayList<>();
    }
}

class SubClass extends SuperClass {
    @Override
    public ArrayList<Integer> func() throws Exception {
        return new ArrayList<>();
    }
}
```

在调用一个方法时，先从本类中查找看是否有对应的方法，如果没有再到父类中查看，看是否从父类继承来。否则就要对参数进行转型，转成父类之后看是否有对应的方法。总的来说，方法调用的优先级为：

- this.func(this)
- super.func(this)
- this.func(super)
- super.func(super)


```java
/*
    A
    |
    B
    |
    C
    |
    D
 */


class A {

    public void show(A obj) {
        System.out.println("A.show(A)");
    }

    public void show(C obj) {
        System.out.println("A.show(C)");
    }
}

class B extends A {

    @Override
    public void show(A obj) {
        System.out.println("B.show(A)");
    }
}

class C extends B {
}

class D extends C {
}
```

```java
public static void main(String[] args) {

    A a = new A();
    B b = new B();
    C c = new C();
    D d = new D();

    // 在 A 中存在 show(A obj)，直接调用
    a.show(a); // A.show(A)
    // 在 A 中不存在 show(B obj)，将 B 转型成其父类 A
    a.show(b); // A.show(A)
    // 在 B 中存在从 A 继承来的 show(C obj)，直接调用
    b.show(c); // A.show(C)
    // 在 B 中不存在 show(D obj)，但是存在从 A 继承来的 show(C obj)，将 D 转型成其父类 C
    b.show(d); // A.show(C)

    // 引用的还是 B 对象，所以 ba 和 b 的调用结果一样
    A ba = new B();
    ba.show(c); // A.show(C)
    ba.show(d); // A.show(C)
}
```

**2. 重载（Overload）**  

存在于同一个类中，指一个方法与已经存在的方法名称上相同，但是参数类型、个数、顺序至少有一个不同。

应该注意的是，返回值不同，其它都相同不算是重载。

```java
class OverloadingExample {
    public void show(int x) {
        System.out.println(x);
    }

    public void show(int x, String y) {
        System.out.println(x + " " + y);
    }
}
```

```java
public static void main(String[] args) {
    OverloadingExample example = new OverloadingExample();
    example.show(1);
    example.show(1, "2");
}
```

# 十、反射

每个类都有一个   **Class**   对象，包含了与类有关的信息。当编译一个新类时，会产生一个同名的 .class 文件，该文件内容保存着 Class 对象。

类加载相当于 Class 对象的加载，类在第一次使用时才动态加载到 JVM 中。也可以使用 `Class.forName("com.mysql.jdbc.Driver")` 这种方式来控制类的加载，该方法会返回一个 Class 对象。

反射可以提供运行时的类信息，并且这个类可以在运行时才加载进来，甚至在编译时期该类的 .class 不存在也可以加载进来。

Class 和 java.lang.reflect 一起对反射提供了支持，java.lang.reflect 类库主要包含了以下三个类：

-   **Field**  ：可以使用 get() 和 set() 方法读取和修改 Field 对象关联的字段；
-   **Method**  ：可以使用 invoke() 方法调用与 Method 对象关联的方法；
-   **Constructor**  ：可以用 Constructor 的 newInstance() 创建新的对象。

**反射的优点：**  

*     **可扩展性**   ：应用程序可以利用全限定名创建可扩展对象的实例，来使用来自外部的用户自定义类。
*     **类浏览器和可视化开发环境**   ：一个类浏览器需要可以枚举类的成员。可视化开发环境（如 IDE）可以从利用反射中可用的类型信息中受益，以帮助程序员编写正确的代码。
*     **调试器和测试工具**   ： 调试器需要能够检查一个类里的私有成员。测试工具可以利用反射来自动地调用类里定义的可被发现的 API 定义，以确保一组测试中有较高的代码覆盖率。

**反射的缺点：**  

尽管反射非常强大，但也不能滥用。如果一个功能可以不用反射完成，那么最好就不用。在我们使用反射技术时，下面几条内容应该牢记于心。

*     **性能开销**   ：反射涉及了动态类型的解析，所以 JVM 无法对这些代码进行优化。因此，反射操作的效率要比那些非反射操作低得多。我们应该避免在经常被执行的代码或对性能要求很高的程序中使用反射。

*     **安全限制**   ：使用反射技术要求程序必须在一个没有安全限制的环境中运行。如果一个程序必须在有安全限制的环境中运行，如 Applet，那么这就是个问题了。

*     **内部暴露**   ：由于反射允许代码执行一些在正常情况下不被允许的操作（比如访问私有的属性和方法），所以使用反射可能会导致意料之外的副作用，这可能导致代码功能失调并破坏可移植性。反射代码破坏了抽象性，因此当平台发生改变的时候，代码的行为就有可能也随着变化。


- [Trail: The Reflection API](https://docs.oracle.com/javase/tutorial/reflect/index.html)
- [深入解析 Java 反射（1）- 基础](http://www.sczyh30.com/posts/Java/java-reflection-1/)

# 十一、异常

Throwable 可以用来表示任何可以作为异常抛出的类，分为两种：  **Error**   和 **Exception**。其中 Error 用来表示 JVM 无法处理的错误，Exception 分为两种：

-   **受检异常**  ：需要用 try...catch... 语句捕获并进行处理，并且可以从异常中恢复；
-   **非受检异常**  ：是程序运行时错误，例如除 0 会引发 Arithmetic Exception，此时程序崩溃并且无法恢复。

<div align="center"> <img src="https://cs-notes-1256109796.cos.ap-guangzhou.myqcloud.com/PPjwP.png" width="600"/> </div><br>

- [Java 入门之异常处理](https://www.cnblogs.com/Blue-Keroro/p/8875898.html)
- [Java Exception Interview Questions and Answers](https://www.journaldev.com/2167/java-exception-interview-questions-and-answersl)

# 十二、泛型

泛型只能是引用类型不能是基本类型。因为基本类型没有地址值，集合里存的都是地址值。如果想存基本数据类型：

```
如果希望向集合ArrayList当中存储基本类型数据，必须使用基本类型对应的“包装类”。

基本类型    包装类（引用类型，包装类都位于java.lang包下）
byte        Byte
short       Short
int         Integer     【特殊】
long        Long
float       Float
double      Double
char        Character   【特殊】
boolean     Boolean

从JDK 1.5+开始，支持自动装箱、自动拆箱。

自动装箱：基本类型 --> 包装类型
自动拆箱：包装类型 --> 基本类型
```

## 泛型概述

<div align="center"><img src="../../pics/8046283a-8d4f-46d4-a716-4c04403568ac.png"width="800px"></img></div>



在前面学习集合时，我们都知道集合中是可以存放任意对象的，只要把对象存储集合后，那么这时他们都会被提升成Object类型。当我们在取出每一个对象，并且进行相应的操作，这时必须采用类型转换。观察下面代码：

~~~java
public class GenericDemo {
	public static void main(String[] args) {
		Collection coll = new ArrayList();
		coll.add("abc");
		coll.add("itcast");
		coll.add(5);//由于集合没有做任何限定，任何类型都可以给其中存放
		Iterator it = coll.iterator();
		while(it.hasNext()){
			//需要打印每个字符串的长度,就要把迭代出来的对象转成String类型
			String str = (String) it.next();
			System.out.println(str.length());
		}
	}
}
~~~

程序在运行时发生了问题**java.lang.ClassCastException**。为什么会发生类型转换异常呢？                                                                                                                                       

由于集合中什么类型的元素都可以存储。导致取出时强转引发运行时 ClassCastException。                                                                                                                                                       

怎么来解决这个问题呢？                                                                                                                                                           

Collection虽然可以存储各种对象，但实际上通常Collection只存储同一类型对象。例如都是存储字符串对象。因此在JDK5之后，新增了**泛型**(**Generic**)语法，让你在设计API时可以指定类或方法支持泛型，这样我们使用API的时候也变得更为简洁，并得到了编译时期的语法检查。

* **泛型**：可以在类或方法中预支地使用未知的类型。

> tips:一般在创建对象时，将未知的类型确定具体的类型。当没有指定泛型时，默认类型为Object类型。

## 使用泛型的好处

* 将运行时期的ClassCastException，转移到了编译时期变成了编译失败。
* 避免了类型强转的麻烦。

通过如下代码体验一下：

~~~java
public class GenericDemo2 {
	public static void main(String[] args) {
        Collection<String> list = new ArrayList<String>();
        list.add("abc");
        list.add("itcast");
        // list.add(5);//当集合明确类型后，存放类型不一致就会编译报错
        // 集合已经明确具体存放的元素类型，那么在使用迭代器的时候，迭代器也同样会知道具体遍历元素类型
        Iterator<String> it = list.iterator();
        while(it.hasNext()){
            String str = it.next();
            //当使用Iterator<String>控制元素类型后，就不需要强转了。获取到的元素直接就是String类型
            System.out.println(str.length());
        }
	}
}
~~~

> tips:泛型是数据类型的一部分，我们将类名与泛型合并一起看做数据类型。

## 定义与使用

泛型，用来灵活地将数据类型应用到不同的类、方法、接口当中。将数据类型作为参数进行传递。

### 定义和使用含有泛型的类

定义格式：

~~~
修饰符 class 类名<代表泛型的变量> {  }
~~~

例如，API中的ArrayList集合：

~~~java
class ArrayList<E>{ 
    public boolean add(E e){ }

    public E get(int index){ }
   	....
}
~~~

使用泛型： 即什么时候确定泛型。

**在创建对象的时候确定泛型**

 例如，`ArrayList<String> list = new ArrayList<String>();`

此时，变量E的值就是String类型,那么我们的类型就可以理解为：

~~~java 
class ArrayList<String>{ 
     public boolean add(String e){ }

     public String get(int index){  }
     ...
}
~~~

再例如，`ArrayList<Integer> list = new ArrayList<Integer>();`

此时，变量E的值就是Integer类型,那么我们的类型就可以理解为：

~~~java
class ArrayList<Integer> { 
     public boolean add(Integer e) { }

     public Integer get(int index) {  }
     ...
}
~~~

举例自定义泛型类

~~~java
public class MyGenericClass<MVP> {
	//没有MVP类型，在这里代表 未知的一种数据类型 未来传递什么就是什么类型
	private MVP mvp;
     
    public void setMVP(MVP mvp) {
        this.mvp = mvp;
    }
     
    public MVP getMVP() {
        return mvp;
    }
}
~~~

使用:

~~~java
public class GenericClassDemo {
  	public static void main(String[] args) {		 
         // 创建一个泛型为String的类
         MyGenericClass<String> my = new MyGenericClass<String>();    	
         // 调用setMVP
         my.setMVP("大胡子登登");
         // 调用getMVP
         String mvp = my.getMVP();
         System.out.println(mvp);
         //创建一个泛型为Integer的类
         MyGenericClass<Integer> my2 = new MyGenericClass<Integer>(); 
         my2.setMVP(123);   	  
         Integer mvp2 = my2.getMVP();
    }
}
~~~

###  含有泛型的方法

定义格式：

~~~
修饰符 <代表泛型的变量> 返回值类型 方法名(参数){  }
~~~

例如，

~~~java
public class MyGenericMethod {	  
    public <MVP> void show(MVP mvp) {
    	System.out.println(mvp.getClass());
    }
    
    public <MVP> MVP show2(MVP mvp) {	
    	return mvp;
    }
}
~~~

使用格式：**调用方法时，确定泛型的类型**

~~~java
public class GenericMethodDemo {
    public static void main(String[] args) {
        // 创建对象
        MyGenericMethod mm = new MyGenericMethod();
        // 演示看方法提示
        mm.show("aaa");
        mm.show(123);
        mm.show(12.45);
    }
}
~~~

### 含有泛型的接口

定义格式：

~~~
修饰符 interface接口名<代表泛型的变量> {  }
~~~

例如，

~~~java
public interface MyGenericInterface<E>{
	public abstract void add(E e);
	
	public abstract E getE();  
}
~~~

使用格式：

**1、定义类时确定泛型的类型**

例如

~~~java
public class MyImp1 implements MyGenericInterface<String> {
	@Override
    public void add(String e) {
        // 省略...
    }

	@Override
	public String getE() {
		return null;
	}
}
~~~

此时，泛型E的值就是String类型。

 **2、始终不确定泛型的类型，直到创建对象时，确定泛型的类型**

 例如

~~~java
public class MyImp2<E> implements MyGenericInterface<E> {
	@Override
	public void add(E e) {
       	 // 省略...
	}

	@Override
	public E getE() {
		return null;
	}
}
~~~

确定泛型：

~~~java
/*
 * 使用
 */
public class GenericInterface {
    public static void main(String[] args) {
        MyImp2<String>  my = new MyImp2<String>();  
        my.add("aa");
    }
}
~~~

## 泛型通配符

当使用泛型类或者接口时，传递的数据中，泛型类型不确定，可以通过通配符<?>表示。但是一旦使用泛型的通配符后，只能使用Object类中的共性方法，集合中元素自身方法无法使用。

#### 通配符基本使用

泛型的通配符:**不知道使用什么类型来接收的时候,此时可以使用?,?表示未知通配符。**

此时只能接受数据,不能往该集合中存储数据（不能创建对象使用，只能作为方法的参数使用）。举个例子：

~~~java
public static void main(String[] args) {
    Collection<Intger> list1 = new ArrayList<Integer>();
    getElement(list1);
    Collection<String> list2 = new ArrayList<String>();
    getElement(list2);
}
public static void getElement(Collection<?> coll){}
//？代表可以接收任意类型
~~~

> tips:泛型不存在继承关系 Collection<Object> list = new ArrayList<String>();这种是错误的。

#### 通配符高级使用----受限泛型

之前设置泛型的时候，实际上是可以任意设置的，只要是类就可以设置。但是在JAVA的泛型中可以指定一个泛型的**上限**和**下限**。

**泛型的上限**：

* **格式**： `类型名称 <? extends 类 > 对象名称`
* **意义**： `只能接收该类型及其子类`

**泛型的下限**：

- **格式**： `类型名称 <? super 类 > 对象名称`
- **意义**： `只能接收该类型及其父类型`

比如：现已知Object类，String 类，Number类，Integer类，其中Number是Integer的父类

~~~java
public class Demo06Generic {
    public static void main(String[] args) {
        Collection<Integer> list1 = new ArrayList<Integer>();
        Collection<String> list2 = new ArrayList<String>();
        Collection<Number> list3 = new ArrayList<Number>();
        Collection<Object> list4 = new ArrayList<Object>();

        getElement1(list1);
        //getElement1(list2);//报错
        getElement1(list3);
        //getElement1(list4);//报错

        //getElement2(list1);//报错
        //getElement2(list2);//报错
        getElement2(list3);
        getElement2(list4);

        /*
            类与类之间的继承关系
            Integer extends Number extends Object
            String extends Object
         */

    }
    // 泛型的上限：此时的泛型?，必须是Number类型或者Number类型的子类
    public static void getElement1(Collection<? extends Number> coll){}
    // 泛型的下限：此时的泛型?，必须是Number类型或者Number类型的父类
    public static void getElement2(Collection<? super Number> coll){}
}
~~~



```java
public class Box<T> {
    // T stands for "Type"
    private T t;
    public void set(T t) { this.t = t; }
    public T get() { return t; }
}
```

- [Java 泛型详解](http://www.importnew.com/24029.html)
- [10 道 Java 泛型面试题](https://cloud.tencent.com/developer/article/1033693)

# 十三、注解

Java 注解是附加在代码中的一些元信息，用于一些工具在编译、运行时进行解析和使用，起到说明、配置的功能。注解不会也不能影响代码的实际逻辑，仅仅起到辅助性的作用。

[注解 Annotation 实现原理与自定义注解例子](https://www.cnblogs.com/acm-bingzi/p/javaAnnotation.html)

# 十四、特性

## Java 各版本的新特性

**New highlights in Java SE 8**  

1. Lambda Expressions
2. Pipelines and Streams
3. Date and Time API
4. Default Methods
5. Type Annotations
6. Nashhorn JavaScript Engine
7. Concurrent Accumulators
8. Parallel operations
9. PermGen Error Removed

**New highlights in Java SE 7**  

1. Strings in Switch Statement
2. Type Inference for Generic Instance Creation
3. Multiple Exception Handling
4. Support for Dynamic Languages
5. Try with Resources
6. Java nio Package
7. Binary Literals, Underscore in literals
8. Diamond Syntax

- [Difference between Java 1.8 and Java 1.7?](http://www.selfgrowth.com/articles/difference-between-java-18-and-java-17)
- [Java 8 特性](http://www.importnew.com/19345.html)

## Java 与 C++ 的区别

- Java 是纯粹的面向对象语言，所有的对象都继承自 java.lang.Object，C++ 为了兼容 C 即支持面向对象也支持面向过程。
- Java 通过虚拟机从而实现跨平台特性，但是 C++ 依赖于特定的平台。
- Java 没有指针，它的引用可以理解为安全指针，而 C++ 具有和 C 一样的指针。
- Java 支持自动垃圾回收，而 C++ 需要手动回收。
- Java 不支持多重继承，只能通过实现多个接口来达到相同目的，而 C++ 支持多重继承。
- Java 不支持操作符重载，虽然可以对两个 String 对象执行加法运算，但是这是语言内置支持的操作，不属于操作符重载，而 C++ 可以。
- Java 的 goto 是保留字，但是不可用，C++ 可以使用 goto。

[What are the main differences between Java and C++?](http://cs-fundamentals.com/tech-interview/java/differences-between-java-and-cpp.php)

## JRE or JDK

- JRE：Java Runtime Environment，Java 运行环境的简称，为 Java 的运行提供了所需的环境。它是一个 JVM 程序，主要包括了 JVM 的标准实现和一些 Java 基本类库。
- JDK：Java Development Kit，Java 开发工具包，提供了 Java 的开发及运行环境。JDK 是 Java 开发的核心，集成了 JRE 以及一些其它的工具，比如编译 Java 源码的编译器 javac 等。

# 十五、内部类

如果一个事物的内部包含另一个事物，那么这就是一个类内部包含另一个类。
例如：身体和心脏的关系。又如：汽车和发动机的关系。

分类：
1. 成员内部类
2. 局部内部类（包含匿名内部类）

## 成员内部类

成员内部类的定义格式：
修饰符 class 外部类名称 {
    修饰符 class 内部类名称 {
        // ...
    }
    // ...
}

注意：内部类用外部类，随意访问；外用内，需要内部类对象。

--------------------------------------------------------------------------------------------------------------------
如何使用成员内部类？有两种方式：

1. 间接方式：在外部类的方法当中，使用内部类；然后main只是调用外部类的方法。
2. 直接方式，公式：
类名称 对象名 = new 类名称();
【外部类名称.内部类名称 对象名 = new 外部类名称().new 内部类名称();】

```java
public class Body { // 外部类

    public class Heart { // 成员内部类

        // 内部类的方法
        public void beat() {
            System.out.println("心脏跳动：蹦蹦蹦！");
            System.out.println("我叫：" + name); // 正确写法！
        }

    }

    // 外部类的成员变量
    private String name;

    // 外部类的方法
    public void methodBody() {
        System.out.println("外部类的方法");
        new Heart().beat();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }
}


public class Demo01InnerClass {

    public static void main(String[] args) {
        Body body = new Body(); // 外部类的对象
        // 通过外部类的对象，调用外部类的方法，里面间接在使用内部类Heart
        body.methodBody();
        System.out.println("=====================");

        // 按照公式写：
        Body.Heart heart = new Body().new Heart();
        heart.beat();
    }

}
```



如果出现了重名现象，那么格式是：外部类名称.this.外部类成员变量名

```java
public class Outer {

    int num = 10; // 外部类的成员变量

    public class Inner /*extends Object*/ {

        int num = 20; // 内部类的成员变量

        public void methodInner() {
            int num = 30; // 内部类方法的局部变量
            System.out.println(num); // 局部变量，就近原则
            System.out.println(this.num); // 内部类的成员变量
            System.out.println(Outer.this.num); // 外部类的成员变量
        }

    }

}
```



## 局部内部类

如果一个类是定义在一个方法内部的，那么这就是一个局部内部类。“局部”：只有当前所属的方法才能使用它，出了这个方法外面就不能用了。

定义格式：

修饰符 class 外部类名称 {

​    修饰符 返回值类型 外部类方法名称(参数列表) {

​        class 局部内部类名称 {

​            // ...
​        }

​    }

}

定义一个类的时候，权限修饰符规则：

1. 外部类：public / (default)
2. 成员内部类：public / protected / (default) / private
3. 局部内部类：什么都不能写

```java
class Outer {

    public void methodOuter() {
        class Inner { // 局部内部类
            final int num = 10;
            public void methodInner() {
                System.out.println(num); // 10
            }
        }

        Inner inner = new Inner();
        inner.methodInner();
    }

}

```



局部内部类，如果希望访问所在方法的局部变量，那么这个局部变量必须是【有效final的】。

备注：从Java 8+开始，只要局部变量事实不变，那么final关键字可以省略。

原因：
1. new出来的对象在堆内存当中。
2. 局部变量是跟着方法走的，在栈内存当中。
3. 方法运行结束之后，立刻出栈，局部变量就会立刻消失。
4. 但是new出来的对象会在堆当中持续存在，直到垃圾回收消失。

```java
public class MyOuter {

    public void methodOuter() {
        int num = 10; // 所在方法的局部变量

        class MyInner {   // 局部内部类生命周期长，局部变量可能早就出栈了，需要留下唯一不变的数据
            public void methodInner() {
                System.out.println(num);
            }
        }
    }

}
```

 

## 匿名内部类

如果接口的实现类（或者是父类的子类）只需要使用唯一的一次，那么这种情况下就可以省略掉该类的定义，而改为使用【匿名内部类】。

匿名内部类的定义格式：

接口名称 对象名 = new 接口名称() {

​    // 覆盖重写所有抽象方法

};

对格式“new 接口名称() {...}”进行解析：
1. new代表创建对象的动作
2. 接口名称就是匿名内部类需要实现哪个接口
3. {...}这才是匿名内部类的内容

另外还要注意几点问题：
1. 匿名内部类，在【创建对象】的时候，只能使用唯一一次。

  如果希望多次创建对象，而且类的内容一样的话，那么就需要使用单独定义的实现类了。

2. 匿名对象，在【调用方法】的时候，只能调用唯一一次。

  如果希望同一个对象，调用多次方法，那么必须给对象起个名字。

3. 匿名内部类是省略了【实现类/子类名称】，但是匿名对象是省略了【对象名称】

  强调：匿名内部类和匿名对象不是一回事！！！

```java
public interface MyInterface {

    void method1(); // 抽象方法

    void method2();

}

public class MyInterfaceImpl implements MyInterface {
    @Override
    public void method1() {
        System.out.println("实现类覆盖重写了方法！111");
    }

    @Override
    public void method2() {
        System.out.println("实现类覆盖重写了方法！222");
    }
}

public class DemoMain {

    public static void main(String[] args) {
//        MyInterface obj = new MyInterfaceImpl();
//        obj.method();

//        MyInterface some = new MyInterface(); // 错误写法！

        // 使用匿名内部类，但不是匿名对象，对象名称就叫objA
        MyInterface objA = new MyInterface() {
            @Override
            public void method1() {
                System.out.println("匿名内部类实现了方法！111-A");
            }

            @Override
            public void method2() {
                System.out.println("匿名内部类实现了方法！222-A");
            }
        };
        objA.method1();
        objA.method2();
        System.out.println("=================");

        // 使用了匿名内部类，而且省略了对象名称，也是匿名对象
        new MyInterface() {
            @Override
            public void method1() {
                System.out.println("匿名内部类实现了方法！111-B");
            }

            @Override
            public void method2() {
                System.out.println("匿名内部类实现了方法！222-B");
            }
        }.method1();
        // 因为匿名对象无法调用第二次方法，所以需要再创建一个匿名内部类的匿名对象
        new MyInterface() {
            @Override
            public void method1() {
                System.out.println("匿名内部类实现了方法！111-B");
            }

            @Override
            public void method2() {
                System.out.println("匿名内部类实现了方法！222-B");
            }
        }.method2();
    }

}
```

PS. 任何一种类型（类/接口）可以作为成员变量类型；接口可以作为返回值或方法的参数

# 十六、时间

java.util.Calendar类:日历类

Calendar类是一个抽象类,里边提供了很多操作日历字段的方法(YEAR、MONTH、DAY_OF_MONTH、HOUR )

Calendar类无法直接创建对象使用,里边有一个静态方法叫getInstance(),该方法返回了Calendar类的子类对象

static Calendar getInstance() 使用默认时区和语言环境获得一个日历。

```java
public class Demo01Calendar {
    public static void main(String[] args) {
        Calendar c = Calendar.getInstance();//多态
        System.out.println(c);
    }

}
```

# 十七、System

`java.lang.System`类中提供了大量的静态方法，可以获取与系统相关的信息或系统级操作，在System类的API文档中，常用的方法有：

- `public static long currentTimeMillis()`：返回以毫秒为单位的当前时间。
- `public static void arraycopy(Object src, int srcPos, Object dest, int destPos, int length)`：将数组中指定的数据拷贝到另一个数组中。

# 参考资料

- Eckel B. Java 编程思想[M]. 机械工业出版社, 2002.
- Bloch J. Effective java[M]. Addison-Wesley Professional, 2017.
