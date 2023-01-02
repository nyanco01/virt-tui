# virt-tui
<img src="https://user-images.githubusercontent.com/112300116/189452942-6abe3c3d-87e6-4c88-ada1-3e9fa54c3900.png" width="600">
virt-tuiはlibvirtや操作VMの状態確認を行えるTUIアプリケーションです。  
virt-tuiはvirt-managerのように簡単にlibvirtを操作できることを目指して作成されています。(現時点ではほとんど出来ていません。)

## 前提導入パッケージ (Prerequisite Installation Package)
#### Ubuntu 20.04 || Ubuntu 22.04
``` bash
sudo apt install qemu-kvm qemu-system libvirt-daemon-system libvirt-daemon libvirt-dev libvirt-clients bridge-utils libosinfo-bin libguestfs-tools virt-top cloud-image-utils virtinst
```  
go version 1.18
[How to Install](https://go.dev/doc/install)

## Install virt-tui
libvirtがインストールできていてユーザーに以下のようにlibvirtグループに属していればインストール可能です。
``` bash
piyo@ubuntu:~$ id
uid=1000(piyo) gid=1000(piyo) groups=1000(piyo),4(adm),24(cdrom),27(sudo),30(dip),46(plugdev),120(lpadmin),119(libvirt)
```

Ubuntu22.04のみ`/etc/libvirt/qemu.conf`を以下のように設定してください。
```
user = "piyo"
```

#### Install  
``` bash
git clone https://github.com/nyanco01/virt-tui
cd virt-tui  
go build
sudo ./virt-tui
```
※`sudo`なしではVMの作成ができません。  

## 備考 (note)
メモリの使用率を正しく表示させる為には、VMを定義しているxmlファイル内のmemballonの項目に`<stats period='1'/>`を追加する必要があります。  
```xml
    <memballoon model='virtio'>
      <stats period='1'/>
      <alias name='balloon0'/>
    </memballoon>
```
