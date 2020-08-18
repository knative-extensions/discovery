Feature: Reconcile ClusterDuckType

    Scenario Outline: Reconciling Smoke Test with generation <generation>.

        Given the following objects:
            """
            apiVersion: discovery.knative.dev/v1alpha1
            kind: ClusterDuckType
            metadata:
              name: zhangchas.duck.knative.dev
              generation: <generation>
            spec:
              selectors:
                - labelSelector: "duck.knative.dev/smoked=true"

              # Names allows us to give a short name to the duck type.
              names:
                name: "Zhangchas"
                plural: "zhangchas"
                singular: "zhangcha"

              # Versions are to allow the definition of a single duck type with multiple
              # versions, useful if the duck type API shape changes.
              versions:
                - name: "v1"
              group: duck.knative.dev
            status:
              observedGeneration: 0
            """

        And a ClusterDuckType reconciler

        When reconciling "zhangchas.duck.knative.dev"

        Then expect status updates:
            """
            apiVersion: discovery.knative.dev/v1alpha1
            kind: ClusterDuckType
            metadata:
              name: zhangchas.duck.knative.dev
              generation: <generation>
            spec:
              selectors:
                - labelSelector: "duck.knative.dev/smoked=true"

              # Names allows us to give a short name to the duck type.
              names:
                name: "Zhangchas"
                plural: "zhangchas"
                singular: "zhangcha"

              # Versions are to allow the definition of a single duck type with multiple
              # versions, useful if the duck type API shape changes.
              versions:
                - name: "v1"
              group: duck.knative.dev
            status:
              observedGeneration: <generation>
              conditions:
              - type: Ready
                status: "True"
            """

        Examples:
            | generation |
            | 0          |
            | 1          |
            | 2          |

#    # -----------------------------------------
#
#    Scenario: Update status.address on spec.serviceName update.
#
#        Given the following objects:
#            """
#            apiVersion: samples.knative.dev/v1alpha1
#            kind: AddressableService
#            metadata:
#              name: rut
#              namespace: ns
#              generation: 2
#            spec:
#              serviceName: webhook
#            status:
#              observedGeneration: 1
#              address:
#                url: http://old-webhook.ns.svc.cluster.local
#              conditions:
#              - type: Ready
#                status: "True"
#            ---
#            apiVersion: v1
#            kind: Service
#            metadata:
#              name: webhook
#              namespace: ns
#            spec:
#              clusterIP: 10.20.30.40
#              ports:
#              - name: http
#                port: 80
#                protocol: TCP
#                targetPort: 8080
#              sessionAffinity: None
#              type: ClusterIP
#            """
#        And an AddressableService reconciler
#
#        When reconciling "ns/rut"
#
#        Then expect status updates:
#            """
#            apiVersion: samples.knative.dev/v1alpha1
#            kind: AddressableService
#            metadata:
#              name: rut
#              namespace: ns
#              generation: 2
#            spec:
#              serviceName: webhook
#            status:
#              observedGeneration: 2
#              address:
#                url: http://webhook.ns.svc.cluster.local
#              conditions:
#              - type: Ready
#                status: "True"
#            """
#        And expect Kubernetes Events:
#            | Type   | Reason                       | Message                                 |
#            | Normal | AddressableServiceReconciled | AddressableService reconciled: "ns/rut" |
#
#    # -----------------------------------------
#
#    Scenario Outline: Reconciling Normally for generation <generation>.
#
#        Given the following objects:
#            """
#            apiVersion: samples.knative.dev/v1alpha1
#            kind: AddressableService
#            metadata:
#              name: rut
#              namespace: ns
#              generation: <generation>
#            spec:
#              serviceName: webhook
#            ---
#            apiVersion: v1
#            kind: Service
#            metadata:
#              name: webhook
#              namespace: ns
#            spec:
#              clusterIP: 10.20.30.40
#              ports:
#              - name: http
#                port: 80
#                protocol: TCP
#                targetPort: 8080
#              sessionAffinity: None
#              type: ClusterIP
#            """
#        And an AddressableService reconciler
#
#        When reconciling "ns/rut"
#
#        Then expect status updates:
#            """
#            apiVersion: samples.knative.dev/v1alpha1
#            kind: AddressableService
#            metadata:
#              name: rut
#              namespace: ns
#              generation: <generation>
#            spec:
#              serviceName: webhook
#            status:
#              observedGeneration: <generation>
#              address:
#                url: http://webhook.ns.svc.cluster.local
#              conditions:
#              - type: Ready
#                status: "True"
#            """
#        And expect Kubernetes Events:
#            | Type   | Reason                       | Message                                 |
#            | Normal | AddressableServiceReconciled | AddressableService reconciled: "ns/rut" |
#
#        Examples:
#            | generation |
#            | 0          |
#            | 1          |
#            | 2          |
