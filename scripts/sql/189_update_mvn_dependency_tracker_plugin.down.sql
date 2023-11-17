DELETE from plugin_step_variable where name='SkipBuildTest';
update plugin_pipeline_script SET script=E'mkdir $HOME/outDTrack                                                   
OutDirDTrack=$HOME/outDTrack                                            
cd /devtroncd/$CheckoutPath                                             
ToUploadBom=YES                                                         
if [ $BuildToolType == "GRADLE" ]                                       
then                                                                    
        apk add gradle                                                  
        gradle cyclonedxBom                                             
        cp build/reports/bom.json $OutDirDTrack/bom.json                
elif [ $BuildToolType == "MAVEN" ]                                      
then                                                                    
        apk add maven                                                   
        mvn install                    
        cp target/bom.json $OutDirDTrack/bom.json                       
else                                                                    
    echo "BUILD_TYPE: $BuildToolType not supported"                     
    ToUploadBom=NO                                                      
fi                                                                      
                                                                        
if [ $ToUploadBom == "YES" ]                                            
then                                                                    
        apk add curl                                                    
    cd $OutDirDTrack                                                    
        curl -v --location --request POST "$DTrackEndpoint/api/v1/bom" \
                --header "accept: application/json" \                   
        --header "X-Api-Key: $DTrackApiKey" \                           
        --form "projectName=$DTrackProjectName" \                       
        --form \"autoCreate=true\" \                                   
        --form "projectVersion=$DTrackProjectVersion" \                 
        --form "bom=@\"bom.json\""                                       
fi'  WHERE id=5;