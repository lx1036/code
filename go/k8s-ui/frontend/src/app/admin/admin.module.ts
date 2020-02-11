import { NgModule } from '@angular/core';
import {CreateNotificationComponent, ListNotificationComponent, NotificationComponent} from './notification/notification.component';
import {SharedModule} from '../shared/shared.module';
import {RouterModule, Routes} from '@angular/router';
import {AdminComponent} from './admin.component';
import {AdminAuthCheckGuard} from './admin-auth-check-guard.service';
import {NavComponent} from './nav.component';
import {OverviewComponent} from './report-form/overview.component';
import {MarkdownModule} from 'ngx-markdown';
import {AppReportFormComponent} from './report-form/app-report-form.component';
import {DeployComponent} from './report-form/deploy.component';
import {AuditLogComponent} from './audit-log.component';
import {ClusterComponent} from './cluster/cluster.component';
import {TrashClusterComponent} from './cluster/trash-cluster.component';
import {AppComponent} from './app/app.component';
import {TrashAppComponent} from './app/trash-app.component';
import {DeploymentComponent} from './deployment/deployment.component';
import {TrashDeploymentComponent} from './deployment/trash-deployment.component';
import {DeploymentTplComponent} from './deployment/deployment-tpl.component';
import {TrashDeploymentTplComponent} from './deployment/trash-deployment-tpl.component';
import {NamespaceComponent} from './namespace/namespace.component';
import {TrashNamespaceComponent} from './namespace/trash-namespace.component';
import {ConfigmapComponent} from './configmap/configmap.component';
import {TrashConfigmapComponent} from './configmap/trash-configmap.component';
import {ConfigmapTplComponent} from './configmap/configmap-tpl.component';
import {TrashConfigmapTplComponent} from './configmap/trash-configmap-tpl.component';
import {TrashCronjobComponent} from './cronjob/trash-cronjob.component';
import {CronjobTplComponent} from './cronjob/cronjob-tpl.component';
import {TrashCronjobTplComponent} from './cronjob/trash-cronjob-tpl.component';
import {UserComponent} from './user/user.component';
import {GroupComponent} from './group/group.component';
import {PermissionComponent} from './permission/permission.component';
import {SecretComponent} from './secret/secret.component';
import {TrashSecretComponent} from './secret/trash-secret.component';
import {SecretTplComponent} from './secret/secret-tpl.component';
import {TrashSecretTplComponent} from './secret/trash-secret-tpl.component';
import {PersistentVolumeClaimComponent} from './pvc/pvc.component';
import {TrashPersistentVolumeClaimComponent} from './pvc/trash-pvc.component';
import {PersistentVolumeClaimTplComponent} from './pvc/pvc-tpl.component';
import {TrashPersistentVolumeClaimTplComponent} from './pvc/trash-pvc-tpl.component';
import {ApikeyComponent} from './apikey/apikey.component';
import {StatefulsetComponent} from './statefulset/statefulset.component';
import {StatefulsetTplComponent} from './statefulset/statefulset-tpl.component';
import {TrashStatefulsetTplComponent} from './statefulset/trash-statefulset-tpl.component';
import {TrashStatefulsetComponent} from './statefulset/trash-statefulset.component';
import {DaemonSetComponent} from './daemonset/daemonset.component';
import {TrashDaemonSetComponent} from './daemonset/trash-daemonset.component';
import {DaemonSetTplComponent} from './daemonset/daemonset-tpl.component';
import {TrashDaemonSetTplComponent} from './daemonset/trash-daemonset-tpl.component';
import {ConfigComponent} from './config/config.component';
import {ConfigSystemComponent} from './config/list-config-system.component';
import {KubernetesDashboardComponent} from './kubernetes/dashboard.component';
import {NodesComponent} from './kubernetes/node/node.component';
import {PersistentVolumeComponent} from './kubernetes/persistent-volume/persistent-volume.component';
import {CreateEditPersistentVolumeComponent} from './kubernetes/persistent-volume/create-edit-persistent-volume.component';
import {IngressComponent} from './ingress/ingress.component';
import {TrashIngressComponent} from './ingress/trash-ingress.component';
import {IngressTplComponent} from './ingress/ingress-tpl.component';
import {TrashIngressTplComponent} from './ingress/trash-ingress-tpl.component';
import {AutoscaleComponent} from './autoscale/autoscale.component';
import {TrashAutoscaleComponent} from './autoscale/trash-autoscale.component';
import {AutoscaleTplComponent} from './autoscale/autoscale-tpl.component';
import {TrashAutoscaleTplComponent} from './autoscale/trash-autoscale-tpl.component';
import {KubeDeploymentComponent} from './kubernetes/deployment/kube-deployment.component';
import {KubeNamespaceComponent} from './kubernetes/namespace/kube-namespace.component';
import {KubePodComponent} from './kubernetes/pod/kube-pod.component';
import {KubeServiceComponent} from './kubernetes/service/kube-service.component';
import {KubeEndpointComponent} from './kubernetes/endpoint/kube-endpoint.component';
import {KubeConfigmapComponent} from './kubernetes/configmap/kube-configmap.component';
import {KubeSecretComponent} from './kubernetes/secret/kube-secret.component';
import {KubeIngressComponent} from './kubernetes/ingress/kube-ingress.component';
import {KubeStatefulsetComponent} from './kubernetes/statefulset/kube-statefulset.component';
import {KubeDaemonsetComponent} from './kubernetes/daemonset/kube-daemonset';
import {KubeCronjobComponent} from './kubernetes/cronjob/kube-cronjob.component';
import {KubeJobComponent} from './kubernetes/job/kube-job.component';
import {KubePvcComponent} from './kubernetes/pvc/kube-pvc.component';
import {KubeReplicasetComponent} from './kubernetes/replicaset/kube-replicaset.component';
import {KubeStorageclassComponent} from './kubernetes/storageclass/kube-storageclass.component';
import {KubeHpaComponent} from './kubernetes/hpa/kube-hpa.component';
import {KubeRoleComponent} from './kubernetes/role/kube-role.component';
import {KubeRolebindingComponent} from './kubernetes/rolebinding/kube-rolebinding.component';
import {KubeServiceaccountComponent} from './kubernetes/serviceaccount/kube-serviceaccount.component';
import {KubeClusterroleComponent} from './kubernetes/clusterrole/kube-clusterrole.component';
import {KubeClusterrolebindingComponent} from './kubernetes/clusterrolebinding/kube-clusterrolebinding.component';
import {KubeCrdComponent} from './kubernetes/crd/kube-crd.component';
import {CronjobComponent} from './cronjob/cronjob.component';


const routes: Routes = [
  {
    path: 'admin',
    component: AdminComponent,
    canActivate: [AdminAuthCheckGuard],
    canActivateChild: [AdminAuthCheckGuard],
    children: [
      {
        path: '',
        pathMatch: 'full',
        redirectTo: 'reportform/overview'
      },
      {path: 'reportform/deploy', component: DeployComponent},
      {path: 'reportform/overview', component: OverviewComponent},
      {path: 'reportform/app', component: AppReportFormComponent},
      {path: 'auditlog', component: AuditLogComponent},
      {path: 'notification', component: NotificationComponent},
      {path: 'cluster', component: ClusterComponent},
      {path: 'cluster/trash', component: TrashClusterComponent},
      {path: 'app', component: AppComponent},
      {path: 'app/trash', component: TrashAppComponent},
      {path: 'deployment', component: DeploymentComponent},
      {path: 'deployment/trash', component: TrashDeploymentComponent},
      {path: 'deployment/tpl', component: DeploymentTplComponent},
      {path: 'deployment/tpl/trash', component: TrashDeploymentTplComponent},
      {path: 'namespace', component: NamespaceComponent},
      {path: 'namespace/trash', component: TrashNamespaceComponent},
      {path: 'configmap', component: ConfigmapComponent},
      {path: 'configmap/trash', component: TrashConfigmapComponent},
      {path: 'configmap/tpl', component: ConfigmapTplComponent},
      {path: 'configmap/tpl/trash', component: TrashConfigmapTplComponent},
      {path: 'cronjob', component: CronjobComponent},
      {path: 'cronjob/trash', component: TrashCronjobComponent},
      {path: 'cronjob/tpl', component: CronjobTplComponent},
      {path: 'cronjob/tpl/trash', component: TrashCronjobTplComponent},
      {path: 'system/user', component: UserComponent},
      {path: 'system/user/:gid', component: UserComponent},
      {path: 'system/group', component: GroupComponent},
      {path: 'system/permission', component: PermissionComponent},
      {path: 'secret', component: SecretComponent},
      {path: 'secret/trash', component: TrashSecretComponent},
      {path: 'secret/tpl', component: SecretTplComponent},
      {path: 'secret/tpl/trash', component: TrashSecretTplComponent},
      {path: 'persistentvolumeclaim', component: PersistentVolumeClaimComponent},
      {path: 'persistentvolumeclaim/trash', component: TrashPersistentVolumeClaimComponent},
      {path: 'persistentvolumeclaim/tpl', component: PersistentVolumeClaimTplComponent},
      {path: 'persistentvolumeclaim/tpl/trash', component: TrashPersistentVolumeClaimTplComponent},
      {path: 'apikey', component: ApikeyComponent},
      {path: 'statefulset', component: StatefulsetComponent},
      {path: 'statefulset/trash', component: TrashStatefulsetComponent},
      {path: 'statefulset/tpl', component: StatefulsetTplComponent},
      {path: 'statefulset/tpl/trash', component: TrashStatefulsetTplComponent},
      {path: 'daemonset', component: DaemonSetComponent},
      {path: 'daemonset/trash', component: TrashDaemonSetComponent},
      {path: 'daemonset/tpl', component: DaemonSetTplComponent},
      {path: 'daemonset/tpl/trash', component: TrashDaemonSetTplComponent},
      {path: 'config/database', component: ConfigComponent},
      {path: 'config/system', component: ConfigSystemComponent},
      {path: 'kubernetes/dashboard', component: KubernetesDashboardComponent},
      {path: 'kubernetes/dashboard/:cluster', component: KubernetesDashboardComponent},
      {path: 'kubernetes/node', component: NodesComponent},
      {path: 'kubernetes/node/:cluster', component: NodesComponent},
      {path: 'kubernetes/persistentvolume', component: PersistentVolumeComponent},
      {path: 'kubernetes/persistentvolume/:cluster', component: PersistentVolumeComponent},
      {path: 'kubernetes/persistentvolume/:cluster/edit/:name', component: CreateEditPersistentVolumeComponent},
      {path: 'kubernetes/persistentvolume/:cluster/edit', component: CreateEditPersistentVolumeComponent},
      {path: 'ingress', component: IngressComponent},
      {path: 'ingress/trash', component: TrashIngressComponent},
      {path: 'ingress/tpl', component: IngressTplComponent},
      {path: 'ingress/tpl/trash', component: TrashIngressTplComponent},
      {path: 'hpa', component: AutoscaleComponent},
      {path: 'hpa/trash', component: TrashAutoscaleComponent},
      {path: 'hpa/tpl', component: AutoscaleTplComponent},
      {path: 'hpa/tpl/trash', component: TrashAutoscaleTplComponent},
      {path: 'kubernetes/deployment', component: KubeDeploymentComponent},
      {path: 'kubernetes/deployment/:cluster', component: KubeDeploymentComponent},
      {path: 'kubernetes/namespace', component: KubeNamespaceComponent},
      {path: 'kubernetes/namespace/:cluster', component: KubeNamespaceComponent},
      {path: 'kubernetes/pod', component: KubePodComponent},
      {path: 'kubernetes/pod/:cluster', component: KubePodComponent},
      {path: 'kubernetes/service', component: KubeServiceComponent},
      {path: 'kubernetes/service/:cluster', component: KubeServiceComponent},
      {path: 'kubernetes/endpoint', component: KubeEndpointComponent},
      {path: 'kubernetes/endpoint/:cluster', component: KubeEndpointComponent},
      {path: 'kubernetes/configmap', component: KubeConfigmapComponent},
      {path: 'kubernetes/configmap/:cluster', component: KubeConfigmapComponent},
      {path: 'kubernetes/secret', component: KubeSecretComponent},
      {path: 'kubernetes/secret/:cluster', component: KubeSecretComponent},
      {path: 'kubernetes/ingress', component: KubeIngressComponent},
      {path: 'kubernetes/ingress/:cluster', component: KubeIngressComponent},
      {path: 'kubernetes/statefulset', component: KubeStatefulsetComponent},
      {path: 'kubernetes/statefulset/:cluster', component: KubeStatefulsetComponent},
      {path: 'kubernetes/daemonset', component: KubeDaemonsetComponent},
      {path: 'kubernetes/daemonset/:cluster', component: KubeDaemonsetComponent},
      {path: 'kubernetes/cronjob', component: KubeCronjobComponent},
      {path: 'kubernetes/cronjob/:cluster', component: KubeCronjobComponent},
      {path: 'kubernetes/job', component: KubeJobComponent},
      {path: 'kubernetes/job/:cluster', component: KubeJobComponent},
      {path: 'kubernetes/persistentvolumeclaim', component: KubePvcComponent},
      {path: 'kubernetes/persistentvolumeclaim/:cluster', component: KubePvcComponent},
      {path: 'kubernetes/replicaset', component: KubeReplicasetComponent},
      {path: 'kubernetes/replicaset/:cluster', component: KubeReplicasetComponent},
      {path: 'kubernetes/storageclass', component: KubeStorageclassComponent},
      {path: 'kubernetes/storageclass/:cluster', component: KubeStorageclassComponent},
      {path: 'kubernetes/horizontalpodautoscaler', component: KubeHpaComponent},
      {path: 'kubernetes/horizontalpodautoscaler/:cluster', component: KubeHpaComponent},
      {path: 'kubernetes/role', component: KubeRoleComponent},
      {path: 'kubernetes/role/:cluster', component: KubeRoleComponent},
      {path: 'kubernetes/rolebinding', component: KubeRolebindingComponent},
      {path: 'kubernetes/rolebinding/:cluster', component: KubeRolebindingComponent},
      {path: 'kubernetes/serviceaccount', component: KubeServiceaccountComponent},
      {path: 'kubernetes/serviceaccount/:cluster', component: KubeServiceaccountComponent},
      {path: 'kubernetes/clusterrole', component: KubeClusterroleComponent},
      {path: 'kubernetes/clusterrole/:cluster', component: KubeClusterroleComponent},
      {path: 'kubernetes/clusterrolebinding', component: KubeClusterrolebindingComponent},
      {path: 'kubernetes/clusterrolebinding/:cluster', component: KubeClusterrolebindingComponent},
      {path: 'kubernetes/customresourcedefinition', component: KubeCrdComponent},
      {path: 'kubernetes/customresourcedefinition/:cluster', component: KubeCrdComponent},
    ]
  }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class AdminRoutingModule {
}

@NgModule({
  declarations: [
    CronjobComponent,
    NotificationComponent,
    CreateNotificationComponent,
    ListNotificationComponent,
    AdminComponent,
    NavComponent,
    OverviewComponent,
  ],
  imports: [
    SharedModule,
    AdminRoutingModule,
    MarkdownModule.forRoot(),
  ],
  providers: [
    AdminAuthCheckGuard,
  ]
})
export class AdminModule { }
