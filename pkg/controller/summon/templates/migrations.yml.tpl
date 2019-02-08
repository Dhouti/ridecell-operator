apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Instance.Name }}-migrations
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: migrations
    app.kubernetes.io/instance: {{ .Instance.Name }}-migrations
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: migration
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/name: migrations
        app.kubernetes.io/instance: {{ .Instance.Name }}-migrations
        app.kubernetes.io/version: {{ .Instance.Spec.Version }}
        app.kubernetes.io/component: migration
        app.kubernetes.io/part-of: {{ .Instance.Name }}
        app.kubernetes.io/managed-by: summon-operator
    spec:
      restartPolicy: Never
      imagePullSecrets:
      - name: pull-secret
      containers:
      - name: default
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command:
        - sh
        - "-c"
        - python manage.py migrate && python manage.py loadflavor etc/summonflavor/{{ .Instance.Spec.SummonFlavor }}.json
        resources:
          requests:
            memory: 1G
            cpu: 500m
          limits:
            memory: 2G
            cpu: 2
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: app-secrets
          mountPath: /etc/secrets
        - name: summonflavor
          mountPath: /etc/summonflavor
      initContainers:
      - name: getflavor
        image: alpine
        command:
        - "sh"
        - "-c"
        - wget -O /etc/summonflavor/{{ .Instance.Spec.SummonFlavor }}.json {{ .Extra.presignedUrl }}
        volumeMounts:
        - name: summonflavor
          mountPath: /etc/summonflavor

      volumes:
        - name: config-volume
          configMap:
            name: {{ .Instance.Name }}-config
        - name: app-secrets
          secret:
            secretName: summon.{{ .Instance.Name }}.app-secrets
        - name: summonflavor
          emtpyDir: {}
