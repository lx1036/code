
# check
# ansible-playbook -i host.ini task.yml --syntax-check -tags="push"

# https://docs.ansible.com/ansible/latest/user_guide/playbooks_best_practices.html
ansible-playbook -i inventories/prd/host.ini task.yml -tags="push"
