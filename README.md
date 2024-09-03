# docker-certs

## Prerequisites:
- make sure you have ansible installed

## How to:
1. clone project
2. Run:
```bash 
ansible-playbook -i inventory.ini --ask-vault-pass --ask-pass -K install_mkcert.yaml
```

