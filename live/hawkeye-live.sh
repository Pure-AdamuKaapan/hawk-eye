orig_dir=$(pwd)
uuid=$(uuidgen)
mkdir $uuid
cd $uuid
for pod in $(kubectl -n kube-system get pods --no-headers -l name=portworx -o wide| awk '{print $1}')
do
mkdir $pod
cd $pod
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl sv pool show > pxctl_sv_pool_show.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt blkid -f > blkid.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt lsblk > lsblk.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt uname -a > uname.out 2>/dev/null
cd ..
done



for pod in $(kubectl -n kube-system get pods --no-headers -l name=portworx -o wide| awk '{print $1}'|head -1)
do
cd $pod
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl cd list-drive > pxctl_cd_list_drive.out 2>/dev/null
kubectl -nkube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl sv kvdb members > pxctl_sv_kvdb_members.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl status > pxctl_status.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl v l > pxctl_v_l.out 2>/dev/null
kubectl -n kube-system exec -it $pod -c portworx -- nsenter --mount=/host_proc/1/ns/mnt pxctl alerts show > pxctl_alerts_show.out 2>/dev/null
cd ..
done

echo $uuid 
root_dir=$(pwd)
cd ../../display/
python3 hawkeye_report.py "$root_dir"
mv $root_dir/index.html $root_dir/$uuid.html
cd $root_dir
$orig_dir/s3 upload $uuid.html
