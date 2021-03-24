Feature: Reconcile ClusterDuckType in a Zoo

    Scenario Outline: Reconciling ClusterDuckType <key>

        Given the following objects (from file):
            | file                     |
            | config/zoo/animals.yaml  |
            | config/zoo/initial.yaml  |
            | config/zoo/clusterroles.yaml  |

        And a ClusterDuckType reconciler

        When reconciling "<key>"

        Then expect status updates (from file):
            | file      |
            | <updated> |

        Examples:
            | key                        | updated                            |
            | ears.zoo.knative.dev       | config/zoo/updated-ears.yaml       |
            | furries.zoo.knative.dev    | config/zoo/updated-furries.yaml    |
            | bills.zoo.knative.dev      | config/zoo/updated-bills.yaml      |
            | swimmers.zoo.knative.dev   | config/zoo/updated-swimmers.yaml   |
