ALTER TABLE ONLY public.image_scan_deploy_info
    ADD CONSTRAINT image_scan_deploy_info_scan_object_meta_id_fkey FOREIGN KEY (scan_object_meta_id) REFERENCES public.image_scan_object_meta(id);
