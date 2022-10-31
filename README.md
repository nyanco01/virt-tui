# virt-tui
<img src="https://user-images.githubusercontent.com/112300116/189452942-6abe3c3d-87e6-4c88-ada1-3e9fa54c3900.png" width="600">
virt-tuiはlibvirtや操作VMの状態確認を行えるTUIアプリケーションです。  
virt-tuiは[virt-manager](https://github.com/virt-manager/virt-manager)のように簡単にlibvirtを操作できることを目指して作成されています。
(現時点ではほとんど出来ていません。)

virt-tui is a TUI application that allows you to check the status of libvirt and operating VMs.  
virt-tui was created with the goal of making it as easy to manipulate libvirt as [virt-manager](https://github.com/virt-manager/virt-manager). (At the moment, it is almost not done.)

## 前提導入パッケージ (Prerequisite Installation Package)
#### Ubuntu 20.04
``` bash
sudo apt install qemu-kvm qemu-system libvirt-daemon-system libvirt-daemon libvirt-dev libvirt-clients bridge-utils libosinfo-bin libguestfs-tools virt-top cloud-image-utils virtinst
```  
go version 1.18
[How to Install](https://go.dev/doc/install)

## Install virt-tui
``` bash
piyo@ubuntu:~$ id
uid=1000(piyo) gid=1000(piyo) groups=1000(piyo),4(adm),24(cdrom),27(sudo),30(dip),46(plugdev),120(lpadmin),131(sambashare)
```
ユーザーにlibvirtの権限を付与します。  
Grant libvirt privileges to users.

``` bash
piyo@ubuntu-hoge:~$ sudo su
root@ubuntu-libvirt3:/home/piyo# cat > /etc/polkit-1/localauthority/50-local.d/50-libvirt.pkla  <<EOF
[Passwordless libvirt access]
Identity=unix-group:piyo
Action=org.libvirt.unix.manage
ResultAny=yes
ResultInactive=yes
ResultActive=yes
EOF
```

Install  
``` bash
git clone https://github.com/nyanco01/virt-tui
cd virt-tui  
go build
sudo ./virt-tui
```
※`sudo`なしではVMの作成ができません。  
VM cannot be created without sudo.

## 備考 (note)
メモリの使用率を正しく表示させる為には、VMを定義しているxmlファイル内のmemballonの項目に`<stats period='1'/>`を追加する必要があります。  
In order to display the memory usage correctly, you need to add an `<stats period='1'/>` to the memballon entry in the xml file that defines the VM. 
```xml
    <memballoon model='virtio'>
      <stats period='1'/>
      <alias name='balloon0'/>
    </memballoon>
```
