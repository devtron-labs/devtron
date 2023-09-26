DROP index image_scan_deploy_info;
CREATE UNIQUE INDEX image_scan_deploy_info_unique ON public.image_scan_deploy_info USING btree (scan_object_meta_id, object_type);
