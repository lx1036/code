




**[stakater/Jamadar](https://github.com/stakater/Jamadar)**


# Problem
Dangling/Redundant resources take a lot of space and memory in a cluster. So we want to delete these unneeded resources depending upon the age and pre-defined annotations. 
e.g. I would like to delete namespaces that were without a specific annotation and are almost a month old and would like to take action whenever that happens.


# Solution
Jamadar is a Kubernetes controller that can poll at configured time intervals and watch for dangling resources that are an 'X' time period old 
and don't have a specific annotation, and will delete them and take corresponding actions.

