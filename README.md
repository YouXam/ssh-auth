# ssh-auth

## 开始

**编译**

```shell
make
```

**查看帮助信息**

```shell
./ssh-auth --help
```

**安装**

添加当前目录到环境变量。

```shell
echo "export PATH=\"`pwd`:\$PATH\"" >> ~/.bashrc
```

## 使用方法

```shell
ssh-auth --version
ssh-auth --help
ssh-auth <command> [<args>] 
```

## 命令

### 添加成员/导入公钥

```shell
ssh-auth user <name> [path [path2 [path3 ...]]]
```

**name**: 成员名

**path, path2, path3, ...**: 公钥路径

### 添加服务器

```shell
ssh-auth server [-p port] [-P] [-i path] [-n name] [user@]hostname
```

**-p**: 服务器端口，默认为 22

**-P**: 使用密码，密码将会**明文**保存

**-i path**: 使用公钥，path 为公钥路径，公钥将会保存

**-n name**: 给服务器设置别名

**user**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址

### 添加指定成员的所有公钥到指定服务器

```shell
ssh-auth copy [-p port] [-P] [-i path] <servername|[username@]hostname> <user> [user2 [user3 ...]]
```

成员和服务器关系将会保存。

**-p**: 服务器端口，默认为 22

**-P**: 使用密码连接服务器，覆盖已保存设置，密码将**不会**保存

**-i path**: 使用公钥，path 为公钥路径，覆盖已保存设置，公钥将**不会**保存

**servername** 服务器别名

**username**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址

**user, user2, user3, ...**: 成员名

### 根据保存的成员和服务器关系重新同步公钥

```shell
ssh-auth sync
```