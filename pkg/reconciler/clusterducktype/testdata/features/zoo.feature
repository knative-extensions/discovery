Feature: Reconcile ClusterDuckType in a Zoo

    Scenario: Reconciling ClusterDuckType ears.zoo.knative.dev

        Given the following objects (from file):
            | file                     |
            | config/zoo/animals.yaml  |
            | config/zoo/initial.yaml  |

        And a ClusterDuckType reconciler

        When reconciling "ears.zoo.knative.dev"

        Then expect status updates (from file):
            | file                      |
            | config/zoo/updated-ears.yaml |
