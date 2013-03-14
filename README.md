CloudStorageTesting
===================

# 使用测试程序
	1、 测速程序在demo目录下
	2、 将测试文件放在demo/data目录，建议将测试文件名命名为该文件的大小，例如1.8m/780k/58m等等
	3、 各存储商的配置文件放在demo/config目录下
	4、 运行测试程序
		./test
	5、 测试结果会在当前demo/文件夹下，以Unix时间的整形值命名

# 增加其他云存储提供商的测试用例
	1、 以华为为例，在src/目录下实现华为的sdk
	2、 将配置文件huawei.conf放在demo/config目录下
	3、 创建 hw 存储对象：

		// HuaWei oss
		if err = loadconfig(&huaweiconf, "config/huawei.conf"); err != nil {
			t.Fatal(err)
		}
		hs = huawei.New(huaweiconf)
		T.AddTestStorage(hs, "huawei", "testbucketname")


	4、 运行测试程序

haha
haha
