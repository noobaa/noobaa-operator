package util

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
)

// APIGroupFilterFunc is a filtering function to limit the amount of server groups we load
type APIGroupFilterFunc func(*metav1.APIGroup) bool

// NoGroupFilter is a filtering function that includes all the server groups
func NoGroupFilter(*metav1.APIGroup) bool { return true }

// FastRESTMapper loads the mapper data from the server with filter and concurrency
// and rediscovers the mapping on meta.NoKindMatchError errors
// See https://github.com/kubernetes-sigs/controller-runtime/issues/537
type FastRESTMapper struct {
	Discovery discovery.DiscoveryInterface
	Filter    APIGroupFilterFunc
	Mapper    meta.RESTMapper
}

// NewFastRESTMapper initializes a FastRESTMapper
func NewFastRESTMapper(dc discovery.DiscoveryInterface, filter APIGroupFilterFunc) meta.RESTMapper {
	if filter == nil {
		filter = NoGroupFilter
	}
	return &FastRESTMapper{
		Discovery: dc,
		Filter:    filter,
		Mapper:    restmapper.NewDiscoveryRESTMapper([]*restmapper.APIGroupResources{}),
	}
}

// Discover reads server groups and then reads each group
// to initialize the mapper groups.
func (m *FastRESTMapper) Discover() error {
	groups, err := m.Discovery.ServerGroups()
	if err != nil {
		return err
	}
	wg := wait.Group{}
	totalCount := 0
	filterCount := 0
	var grs []*restmapper.APIGroupResources
	for _, group := range groups.Groups {
		filtered := m.Filter(&group)
		logrus.Tracef("Group: %s %v", group.Name, filtered)
		totalCount++
		if !filtered {
			continue
		}
		filterCount++
		gr := &restmapper.APIGroupResources{
			Group:              group,
			VersionedResources: make(map[string][]metav1.APIResource),
		}
		grs = append(grs, gr)
		wg.Start(func() { LogError(m.DiscoverGroup(gr)) })
	}
	wg.Wait()
	logrus.Tracef("Filtered %d/%d", filterCount, totalCount)
	m.Mapper = restmapper.NewDiscoveryRESTMapper(grs)
	return nil
}

// DiscoverGroup reads the server group
func (m *FastRESTMapper) DiscoverGroup(gr *restmapper.APIGroupResources) error {
	var errResult error
	for _, version := range gr.Group.Versions {
		resources, err := m.Discovery.ServerResourcesForGroupVersion(version.GroupVersion)
		if err != nil {
			errResult = err
		} else {
			gr.VersionedResources[version.Version] = resources.APIResources
		}
	}
	return errResult
}

// DiscoverOnError check if the error is NoMatchError and calls discover
func (m *FastRESTMapper) DiscoverOnError(err error) bool {
	if meta.IsNoMatchError(err) {
		LogError(m.Discover())
		return true
	}
	return false
}

// KindFor implements Mapper.KindFor
func (m *FastRESTMapper) KindFor(resource schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	res, err := m.Mapper.KindFor(resource)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.KindFor(resource)
	}
	return res, err
}

// KindsFor implements Mapper.KindsFor
func (m *FastRESTMapper) KindsFor(resource schema.GroupVersionResource) ([]schema.GroupVersionKind, error) {
	res, err := m.Mapper.KindsFor(resource)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.KindsFor(resource)
	}
	return res, err
}

// ResourceFor implements Mapper.ResourceFor
func (m *FastRESTMapper) ResourceFor(input schema.GroupVersionResource) (schema.GroupVersionResource, error) {
	res, err := m.Mapper.ResourceFor(input)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.ResourceFor(input)
	}
	return res, err
}

// ResourcesFor implements Mapper.ResourcesFor
func (m *FastRESTMapper) ResourcesFor(input schema.GroupVersionResource) ([]schema.GroupVersionResource, error) {
	res, err := m.Mapper.ResourcesFor(input)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.ResourcesFor(input)
	}
	return res, err
}

// RESTMapping implements Mapper.RESTMapping
func (m *FastRESTMapper) RESTMapping(gk schema.GroupKind, versions ...string) (*meta.RESTMapping, error) {
	res, err := m.Mapper.RESTMapping(gk, versions...)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.RESTMapping(gk, versions...)
	}
	return res, err
}

// RESTMappings implements Mapper.RESTMappings
func (m *FastRESTMapper) RESTMappings(gk schema.GroupKind, versions ...string) ([]*meta.RESTMapping, error) {
	res, err := m.Mapper.RESTMappings(gk, versions...)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.RESTMappings(gk, versions...)
	}
	return res, err
}

// ResourceSingularizer implements Mapper.ResourceSingularizer
func (m *FastRESTMapper) ResourceSingularizer(resource string) (string, error) {
	res, err := m.Mapper.ResourceSingularizer(resource)
	if m.DiscoverOnError(err) {
		res, err = m.Mapper.ResourceSingularizer(resource)
	}
	return res, err
}
