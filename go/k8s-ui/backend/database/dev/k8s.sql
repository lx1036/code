# [ORM]2020/01/04 21:40:30 unsupport orm tag namespace
-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.App`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `app` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `namespace_id` bigint NOT NULL,
    `meta_data` longtext NOT NULL,
    `description` varchar(512),
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `app_name` ON `app` (`name`);
CREATE INDEX `app_namespace_id` ON `app` (`namespace_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.APIKey`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `api_key` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `token` longtext NOT NULL,
    `type` integer NOT NULL DEFAULT 0 ,
    `resource_id` bigint,
    `group_id` bigint,
    `description` varchar(512),
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `expire_in` bigint NOT NULL DEFAULT 0 ,
    `deleted` bool NOT NULL DEFAULT false ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `api_key_name` ON `api_key` (`name`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.AppStarred`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `app_starred` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `app_id` bigint NOT NULL,
    `user_id` bigint NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `app_starred_app_id` ON `app_starred` (`app_id`);
CREATE INDEX `app_starred_user_id` ON `app_starred` (`user_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.AppUser`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `app_user` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `app_id` bigint NOT NULL,
    `user_id` bigint NOT NULL,
    `group_id` bigint NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `app_user_app_id` ON `app_user` (`app_id`);
CREATE INDEX `app_user_user_id` ON `app_user` (`user_id`);
CREATE INDEX `app_user_group_id` ON `app_user` (`group_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.AuditLog`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `audit_log` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `subject_id` bigint NOT NULL DEFAULT 0 ,
    `log_type` varchar(128) NOT NULL DEFAULT '' ,
    `log_level` varchar(128) NOT NULL DEFAULT '' ,
    `action` varchar(255) NOT NULL DEFAULT '' ,
    `message` longtext,
    `user_ip` varchar(200) NOT NULL DEFAULT '' ,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `create_time` datetime
) ENGINE=InnoDB;
CREATE INDEX `audit_log_log_type` ON `audit_log` (`log_type`);
CREATE INDEX `audit_log_log_level` ON `audit_log` (`log_level`);
CREATE INDEX `audit_log_action` ON `audit_log` (`action`);
CREATE INDEX `audit_log_user` ON `audit_log` (`user`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Cluster`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `cluster` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `displayname` varchar(512),
    `meta_data` longtext,
    `master` varchar(128) NOT NULL DEFAULT '' ,
    `kube_config` longtext,
    `description` varchar(512),
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false ,
    `status` integer NOT NULL DEFAULT 0 
) ENGINE=InnoDB;

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Charge`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `charge` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `namespace` varchar(1024) NOT NULL DEFAULT '' ,
    `app` varchar(128) NOT NULL DEFAULT '' ,
    `name` varchar(1024) NOT NULL DEFAULT '' ,
    `type` varchar(128) NOT NULL DEFAULT '' ,
    `unit_price` numeric(12, 4) NOT NULL DEFAULT 0 ,
    `quantity` integer NOT NULL DEFAULT 0 ,
    `amount` numeric(12, 4) NOT NULL DEFAULT 0 ,
    `resource_name` varchar(1024) NOT NULL DEFAULT '' ,
    `start_time` datetime NOT NULL,
    `create_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `charge_app` ON `charge` (`app`);
CREATE INDEX `charge_type` ON `charge` (`type`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Config`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `config` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(256) NOT NULL DEFAULT '' ,
    `value` longtext NOT NULL
) ENGINE=InnoDB;

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.ConfigMap`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `config_map` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `config_map_app_id` ON `config_map` (`app_id`);
CREATE INDEX `config_map_order_id` ON `config_map` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.ConfigMapTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `config_map_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `config_map_id` bigint NOT NULL,
    `meta_data` longtext NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `config_map_template_config_map_id` ON `config_map_template` (`config_map_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Cronjob`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `cronjob` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `cronjob_app_id` ON `cronjob` (`app_id`);
CREATE INDEX `cronjob_order_id` ON `cronjob` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.CronjobTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `cronjob_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `cronjob_id` bigint NOT NULL,
    `meta_data` longtext NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `cronjob_template_cronjob_id` ON `cronjob_template` (`cronjob_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.CustomLink`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `custom_link` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `namespace` varchar(255) NOT NULL DEFAULT '' ,
    `link_type` varchar(255) NOT NULL DEFAULT '' ,
    `url` varchar(255) NOT NULL DEFAULT '' ,
    `add_param` bool NOT NULL DEFAULT false ,
    `params` varchar(255),
    `deleted` bool NOT NULL DEFAULT false ,
    `status` bool NOT NULL DEFAULT true 
) ENGINE=InnoDB;
CREATE INDEX `custom_link_namespace` ON `custom_link` (`namespace`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.DaemonSet`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `daemon_set` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `daemon_set_app_id` ON `daemon_set` (`app_id`);
CREATE INDEX `daemon_set_order_id` ON `daemon_set` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.DaemonSetTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `daemon_set_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `daemon_set_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `daemon_set_template_daemon_set_id` ON `daemon_set_template` (`daemon_set_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Deployment`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `deployment` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `deployment_app_id` ON `deployment` (`app_id`);
CREATE INDEX `deployment_order_id` ON `deployment` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.DeploymentTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `deployment_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `deployment_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `deployment_template_deployment_id` ON `deployment_template` (`deployment_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Group`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `group` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(200) NOT NULL DEFAULT '' ,
    `comment` longtext NOT NULL,
    `type` integer NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `group_name` ON `group` (`name`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.HPA`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `hpa` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(255) NOT NULL DEFAULT '' ,
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `hpa_app_id` ON `hpa` (`app_id`);
CREATE INDEX `hpa_order_id` ON `hpa` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.HPATemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `hpa_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `hpa_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `hpa_template_hpa_id` ON `hpa_template` (`hpa_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Ingress`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `ingress` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(255) NOT NULL DEFAULT '' ,
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `ingress_app_id` ON `ingress` (`app_id`);
CREATE INDEX `ingress_order_id` ON `ingress` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.IngressTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `ingress_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `ingress_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `ingress_template_ingress_id` ON `ingress_template` (`ingress_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Invoice`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `invoice` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `namespace` varchar(1024) NOT NULL DEFAULT '' ,
    `app` varchar(128) NOT NULL DEFAULT '' ,
    `amount` numeric(12, 4) NOT NULL DEFAULT 0 ,
    `start_date` datetime NOT NULL,
    `end_date` datetime NOT NULL,
    `bill_date` datetime NOT NULL,
    `create_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `invoice_app` ON `invoice` (`app`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.LinkType`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `link_type` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `type_name` varchar(255) NOT NULL DEFAULT ''  UNIQUE,
    `displayname` varchar(255) NOT NULL DEFAULT '' ,
    `default_url` varchar(255) NOT NULL DEFAULT '' ,
    `param_list` varchar(255),
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Namespace`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `namespace` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `kube_namespace` varchar(128) NOT NULL DEFAULT '' ,
    `meta_data` longtext NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `namespace_kube_namespace` ON `namespace` (`kube_namespace`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.NamespaceUser`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `namespace_user` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `namespace_id` bigint NOT NULL,
    `user_id` bigint NOT NULL,
    `group_id` bigint NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `namespace_user_namespace_id` ON `namespace_user` (`namespace_id`);
CREATE INDEX `namespace_user_user_id` ON `namespace_user` (`user_id`);
CREATE INDEX `namespace_user_group_id` ON `namespace_user` (`group_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Notification`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `notification` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `type` varchar(128) NOT NULL DEFAULT '' ,
    `title` varchar(2000) NOT NULL DEFAULT '' ,
    `message` longtext NOT NULL,
    `from_user_id` bigint NOT NULL,
    `level` integer NOT NULL DEFAULT 0 ,
    `is_published` bool NOT NULL DEFAULT false ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `notification_type` ON `notification` (`type`);
CREATE INDEX `notification_from_user_id` ON `notification` (`from_user_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.NotificationLog`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `notification_log` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `user_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `is_readed` bool NOT NULL DEFAULT false ,
    `notification_id` bigint NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `notification_log_notification_id` ON `notification_log` (`notification_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Permission`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `permission` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(200) NOT NULL DEFAULT '' ,
    `comment` longtext NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `permission_name` ON `permission` (`name`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.PersistentVolumeClaim`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `persistent_volume_claim` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `persistent_volume_claim_app_id` ON `persistent_volume_claim` (`app_id`);
CREATE INDEX `persistent_volume_claim_order_id` ON `persistent_volume_claim` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.PersistentVolumeClaimTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `persistent_volume_claim_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `persistent_volume_claim_id` bigint NOT NULL,
    `meta_data` longtext NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `persistent_volume_claim_template_persistent_volume_claim_id` ON `persistent_volume_claim_template` (`persistent_volume_claim_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.PublishHistory`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `publish_history` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `type` integer NOT NULL DEFAULT 0 ,
    `resource_id` bigint NOT NULL DEFAULT 0 ,
    `resource_name` varchar(128) NOT NULL DEFAULT '' ,
    `template_id` bigint NOT NULL DEFAULT 0 ,
    `cluster` varchar(128) NOT NULL DEFAULT '' ,
    `status` integer NOT NULL DEFAULT 0 ,
    `message` longtext NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL
) ENGINE=InnoDB;
CREATE INDEX `publish_history_type` ON `publish_history` (`type`);
CREATE INDEX `publish_history_resource_id` ON `publish_history` (`resource_id`);
CREATE INDEX `publish_history_template_id` ON `publish_history` (`template_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.PublishStatus`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `publish_status` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `type` integer NOT NULL DEFAULT 0 ,
    `resource_id` bigint NOT NULL DEFAULT 0 ,
    `template_id` bigint NOT NULL DEFAULT 0 ,
    `cluster` varchar(128) NOT NULL DEFAULT '' 
) ENGINE=InnoDB;
CREATE INDEX `publish_status_type` ON `publish_status` (`type`);
CREATE INDEX `publish_status_resource_id` ON `publish_status` (`resource_id`);
CREATE INDEX `publish_status_template_id` ON `publish_status` (`template_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Secret`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `secret` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `secret_app_id` ON `secret` (`app_id`);
CREATE INDEX `secret_order_id` ON `secret` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.SecretTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `secret_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `secret_map_id` bigint NOT NULL,
    `meta_data` longtext NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `secret_template_secret_map_id` ON `secret_template` (`secret_map_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Service`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `service` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `service_app_id` ON `service` (`app_id`);
CREATE INDEX `service_order_id` ON `service` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.ServiceTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `service_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `service_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `service_template_service_id` ON `service_template` (`service_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.Statefulset`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `statefulset` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT ''  UNIQUE,
    `meta_data` longtext NOT NULL,
    `app_id` bigint NOT NULL,
    `description` varchar(512),
    `order_id` bigint NOT NULL DEFAULT 0 ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `statefulset_app_id` ON `statefulset` (`app_id`);
CREATE INDEX `statefulset_order_id` ON `statefulset` (`order_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.StatefulsetTemplate`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `statefulset_template` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `template` longtext NOT NULL,
    `statefulset_id` bigint NOT NULL,
    `description` varchar(512) NOT NULL DEFAULT '' ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `statefulset_template_statefulset_id` ON `statefulset_template` (`statefulset_id`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.User`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `user` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(200) NOT NULL DEFAULT ''  UNIQUE,
    `password` varchar(255) NOT NULL DEFAULT '' ,
    `salt` varchar(32) NOT NULL DEFAULT '' ,
    `email` varchar(200) NOT NULL DEFAULT '' ,
    `display` varchar(200) NOT NULL DEFAULT '' ,
    `comment` longtext NOT NULL,
    `type` integer NOT NULL DEFAULT 0 ,
    `admin` bool NOT NULL DEFAULT False ,
    `last_login` datetime NOT NULL,
    `last_ip` varchar(200) NOT NULL DEFAULT '' ,
    `deleted` bool NOT NULL DEFAULT false ,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL
) ENGINE=InnoDB;

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.WebHook`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `web_hook` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `name` varchar(128) NOT NULL DEFAULT '' ,
    `scope` bigint NOT NULL DEFAULT 0 ,
    `object_id` bigint NOT NULL DEFAULT 0 ,
    `url` varchar(512),
    `secret` varchar(512),
    `events` longtext NOT NULL,
    `create_time` datetime NOT NULL,
    `update_time` datetime NOT NULL,
    `user` varchar(128) NOT NULL DEFAULT '' ,
    `enabled` bool NOT NULL DEFAULT false 
) ENGINE=InnoDB;
CREATE INDEX `web_hook_name` ON `web_hook` (`name`);

-- --------------------------------------------------
--  Table Structure for `k8s-lx1036/k8s-ui/backend/models.GroupPermissions`
-- --------------------------------------------------
CREATE TABLE IF NOT EXISTS `group_permissions` (
    `id` bigint AUTO_INCREMENT NOT NULL PRIMARY KEY,
    `group_id` bigint NOT NULL,
    `permission_id` bigint NOT NULL
) ENGINE=InnoDB;
