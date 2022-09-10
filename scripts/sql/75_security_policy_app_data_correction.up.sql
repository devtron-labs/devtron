--
-- Name: cve_policy_control_script it basically updated the duplicate records of app_id, env_id,severity to delete state except for the latest one.
--

update cve_policy_control set deleted = 't' where id IN (
    select main1.id from cve_policy_control main1
                             INNER JOIN
                         (select env_id,severity,max(main.id) as max_id,count(main.id) from cve_policy_control as main where app_id is null and global='f' and deleted='f' group by env_id,severity having count(*) > 1
                         ) AS main2
                         ON main1.env_id = main2.env_id and main1.severity=main2.severity and main1.id != main2.max_id where main1.app_id is null and deleted='f' and global='f'
)



