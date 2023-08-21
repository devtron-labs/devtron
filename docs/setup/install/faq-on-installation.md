
# FAQ

<details>
  <summary>1. How will I know when the installation is finished?</summary>
  
  Run the following command to check the status of the installation:
  
  ```bash
  kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
  ```

  The above command will print `Applied` once the installation process is complete. The installation process could take up to 30 minutes. 
</details>

<details>
  <summary>2. How do I track the progress of the installation?</summary>

  Run the following command to check the logs of the Pod:

  ```bash
  pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod
  ```
</details>

<details>
  <summary>3. How can I restart the installation if the Devtron installer logs contain an error?</summary>

  First run the below command to clean up components installed by Devtron installer:

  ```bash
  cd devtron-installation-script/
  kubectl delete -n devtroncd -f yamls/
  kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'
  ```

  Next, [install Devtron](./install-devtron.md)
</details>

<details>
  <summary>4. What's the purpose of 'Login as administrator' option on the login page?</summary>
  When you install Devtron for the first time, it creates a default admin user and password (with unrestricted access to Devtron). You can use that credentials to log in as an administrator. After the initial login, we recommend you set up any SSO service like Google, GitHub, etc., and then add other users (including yourself). Subsequently, all the users can use the same SSO (let's say, GitHub) to log in to Devtron's dashboard.
</details>

<details>
  <summary>5. Why can't I see the GitOps option in the Global Configurations?</summary>
  If you intend to use GitOps (Argo CD) but unable to see the GitOps settings in 'Global Configurations', chances are that you installed Devtron without the Argo CD module. You can go to 'Devtron Stack Manager' and install GitOps (Argo CD). Post successful installation, the 'GitOps' settings will be available.
</details>

<details>
  <summary>6. Why can't I see the Security settings in the left sidebar?</summary>
  The 'Security' settings primarily exists for two things: one is the scanning of image build generated during the CI process and the other is to create security policies. You can go to 'Devtron Stack Manager' and install either Trivy or Clair (according to your preference). Post successful installation, the 'Security' settings will be available to you in the sidebar.
</details>


Still facing issues, please reach out to us on [Discord](https://discord.gg/jsRG5qx2gp).
