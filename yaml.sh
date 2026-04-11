cat > all.yaml <<'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: realtime-config
  namespace: crash-test
data:
  TZ: "Asia/Shanghai"
  FRONTEND_LISTEN_ADDR: ":8080"
  GAME_BACKEND_URL: "http://realtime-api:18080"
  APP_ENV: "prod"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-frontend
  namespace: crash-test
spec:
  replicas: 2
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: realtime-frontend
  template:
    metadata:
      labels:
        app: realtime-frontend
    spec:
      containers:
        - name: realtime-frontend
          image: flxu/realtime_game:frontend-latest
          imagePullPolicy: Always
          ports:
            - containerPort: 8080
              name: http
          envFrom:
            - configMapRef:
                name: realtime-config
          readinessProbe:
            httpGet:
              path: /
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: "100m"
              memory: "128Mi"
            limits:
              cpu: "500m"
              memory: "512Mi"
---
apiVersion: v1
kind: Service
metadata:
  name: realtime-frontend
  namespace: crash-test
spec:
  selector:
    app: realtime-frontend
  ports:
    - name: http
      port: 80
      targetPort: 8080
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-api
  namespace: crash-test
spec:
  replicas: 2
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: realtime-api
  template:
    metadata:
      labels:
        app: realtime-api
    spec:
      containers:
        - name: realtime-api
          image: flxu/realtime_game:api-latest
          imagePullPolicy: Always
          ports:
            - containerPort: 18080
              name: http
          envFrom:
            - configMapRef:
                name: realtime-config
          readinessProbe:
            tcpSocket:
              port: 18080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            tcpSocket:
              port: 18080
            initialDelaySeconds: 15
            periodSeconds: 20
          resources:
            requests:
              cpu: "200m"
              memory: "256Mi"
            limits:
              cpu: "1000m"
              memory: "1Gi"
---
apiVersion: v1
kind: Service
metadata:
  name: realtime-api
  namespace: crash-test
spec:
  selector:
    app: realtime-api
  ports:
    - name: http
      port: 18080
      targetPort: 18080
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: realtime-worker
  namespace: crash-test
spec:
  replicas: 1
  revisionHistoryLimit: 3
  selector:
    matchLabels:
      app: realtime-worker
  template:
    metadata:
      labels:
        app: realtime-worker
    spec:
      containers:
        - name: realtime-worker
          image: flxu/realtime_game:worker-latest
          imagePullPolicy: Always
          envFrom:
            - configMapRef:
                name: realtime-config
          resources:
            requests:
              cpu: "200m"
              memory: "256Mi"
            limits:
              cpu: "1000m"
              memory: "1Gi"
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: realtime-frontend
  namespace: crash-test
  annotations:
    nginx.ingress.kubernetes.io/proxy-body-size: "20m"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "120"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "120"
spec:
  ingressClassName: nginx
  rules:
    - host: realtime.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: realtime-frontend
                port:
                  number: 80
EOF