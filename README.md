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
ssh-auth user add name [path [path2 [path3 ...]]]
```

**name**: 成员名

**path, path2, path3, ...**: 公钥路径

### 显示所有成员

```shell
ssh-auth user show
```

### 删除成员

```shell
ssh-auth user rm name [name2 [name3 ...]]]
```

**name, name2, name3, ...**: 成员名

### 添加服务器

```shell
ssh-auth server add [-p port] [-P] [-i path] [-n name] [user@]hostname
```

**-p**: 服务器端口，默认为 22

**-P**: 使用密码，密码将会**明文**保存

**-i path**: 使用私钥，path 为私钥路径，私钥将会**明文**保存

**-n name**: 给服务器设置别名

**user**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址

### 显示所有服务器

```shell
ssh-auth server show
```

### 删除服务器

```shell
ssh-auth server rm <servername|[username@]hostname> [servername|[username@]hostname] ...
```

**servername** 服务器别名

**username**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址


### 授权成员登录服务器

将会上传该成员的所有公钥到该服务器，成员和服务器关系将会保存。

```shell
ssh-auth auth add [-p port] [-P] [-i path] <servername|[username@]hostname> <user> [user2 [user3 ...]]
```

**-p**: 服务器端口，默认为 22

**-P**: 使用密码连接服务器，覆盖已保存设置，密码将**不会**保存

**-i path**: 使用私钥，path 为私钥路径，覆盖已保存设置，私钥将**不会**保存

**servername** 服务器别名

**username**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址

**user, user2, user3, ...**: 成员名

### 显示所有授权

```shell
ssh-auth auth show
```

### 删除某授权

当且仅当某服务器的某个公钥没有任何成员在使用时，会尝试连接服务器并删除该公钥。

```shell
ssh-auth auth rm id [id2 [id3 ...]]]
```

**id, id2, id3**: 授权 id，可通过 `ssh-auth auth show` 查看。

### 根据保存的成员和服务器关系重新同步公钥

```shell
ssh-auth sync [servername|[username@]hostname] [servername|[username@]hostname] ...
```

当不提供服务器列表时，默认重新同步所有服务器的公钥。

**servername** 服务器别名

**username**: 服务器用户名，默认为当前用户名

**hostname**: 主机名或 IP 地址
