//  Copyright Istio Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package remote

import (
	"testing"

	"istio.io/istio/pkg/test/framework/components/environment/kube"

	"istio.io/istio/tests/integration/multicluster"

	"istio.io/istio/pkg/test/framework"
	"istio.io/istio/pkg/test/framework/components/istio"
	"istio.io/istio/pkg/test/framework/components/namespace"
	"istio.io/istio/pkg/test/framework/components/pilot"
	"istio.io/istio/pkg/test/framework/label"
	"istio.io/istio/pkg/test/framework/resource"
)

var (
	ist                              istio.Instance
	pilots                           []pilot.Instance
	clusterLocalNS, mcReachabilityNS namespace.Instance
	controlPlaneValues               string
)

func TestMain(m *testing.M) {
	framework.
		NewSuite(m).
		Label(label.Multicluster).
		RequireMinClusters(2).
		Setup(multicluster.Setup(&controlPlaneValues, &clusterLocalNS, &mcReachabilityNS)).
		Setup(kube.Setup(func(s *kube.Settings) {
			// Make all clusters use the same control plane
			s.ControlPlaneTopology = make(map[resource.ClusterIndex]resource.ClusterIndex)
			primaryCluster := resource.ClusterIndex(0)
			for i := 0; i < len(s.KubeConfig); i++ {
				s.ControlPlaneTopology[resource.ClusterIndex(i)] = primaryCluster
			}
		})).
		Setup(istio.Setup(&ist, func(cfg *istio.Config) {
			// Set the control plane values on the config.
			cfg.ControlPlaneValues = controlPlaneValues
		})).
		Setup(func(ctx resource.Context) (err error) {
			pilots = make([]pilot.Instance, len(ctx.Environment().Clusters()))
			for i, cluster := range ctx.Environment().Clusters() {
				if pilots[i], err = pilot.New(ctx, pilot.Config{
					Cluster: cluster,
				}); err != nil {
					return err
				}
			}
			return nil
		}).
		Run()
}

func TestMulticlusterReachability(t *testing.T) {
	multicluster.ReachabilityTest(t, mcReachabilityNS, "installation.multicluster.remote")
}

func TestClusterLocalService(t *testing.T) {
	multicluster.ClusterLocalTest(t, clusterLocalNS, "installation.multicluster.remote")
}

func TestTelemetry(t *testing.T) {
	multicluster.TelemetryTest(t, mcReachabilityNS, "installation.multicluster.remote")
}
