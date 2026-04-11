# realtime_game

make release-all

kubectl get pod -n realtime
kubectl get svc -n realtime
kubectl get ingress -n realtime
kubectl describe pod -n realtime
kubectl logs -n realtime deploy/realtime-api
kubectl logs -n realtime deploy/realtime-worker
kubectl logs -n realtime deploy/realtime-frontend



