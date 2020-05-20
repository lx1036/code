


**[stakater/Chowkidar](https://github.com/stakater/Chowkidar)**

# Problem
We would like to watch for relevant events happening inside kubernetes and then perform actions depending upon the criteria. 
e.g. I would like to get a slack notification when a pod is submitted without requests & limits.

# Solution
Chowkidar allows you to have multiple controllers that will continuously watch types in all the namespaces and automatically perform any actions given in the yaml file. 
With this, you can easily check for any criteria on your Pods/other types and take corresponding actions.
