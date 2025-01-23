package eventarc_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-google/google/acctest"
	"github.com/hashicorp/terraform-provider-google/google/envvar"
	"github.com/hashicorp/terraform-provider-google/google/tpgresource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEventarcChannel_basic(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"region":        envvar.GetTestRegionFromEnv(),
		"project_name":  envvar.GetTestProjectFromEnv(),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventarcChannel_basic(context),
			},
			{
				ResourceName:      "google_eventarc_channel.primary",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEventarcChannel_cryptoKeyUpdate(t *testing.T) {
	t.Parallel()

	region := envvar.GetTestRegionFromEnv()
	key1 := acctest.BootstrapKMSKeyWithPurposeInLocationAndName(t, "ENCRYPT_DECRYPT", region, "tf-bootstrap-eventarc-channel-key1")
	key2 := acctest.BootstrapKMSKeyWithPurposeInLocationAndName(t, "ENCRYPT_DECRYPT", region, "tf-bootstrap-eventarc-channel-key2")

	context := map[string]interface{}{
		"region":        region,
		"project_name":  envvar.GetTestProjectFromEnv(),
		"key_ring":      tpgresource.GetResourceNameFromSelfLink(key1.KeyRing.Name),
		"key1":          tpgresource.GetResourceNameFromSelfLink(key1.CryptoKey.Name),
		"key2":          tpgresource.GetResourceNameFromSelfLink(key2.CryptoKey.Name),
		"random_suffix": acctest.RandString(t, 10),
	}

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.AccTestPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories(t),
		Steps: []resource.TestStep{
			{
				Config: testAccEventarcChannel_setCryptoKey(context),
			},
			{
				ResourceName:      "google_eventarc_channel.primary",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEventarcChannel_cryptoKeyUpdate(context),
			},
			{
				ResourceName:      "google_eventarc_channel.primary",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccEventarcChannel_basic(context map[string]interface{}) string {
	return acctest.Nprintf(`
data "google_project" "test_project" {
	project_id  = "%{project_name}"
}

resource "google_eventarc_channel" "primary" {
	location = "%{region}"
	name     = "tf-test-name%{random_suffix}"
	third_party_provider = "projects/${data.google_project.test_project.project_id}/locations/%{region}/providers/datadog"
}
`, context)
}

func testAccEventarcChannel_setCryptoKey(context map[string]interface{}) string {
	return acctest.Nprintf(`
data "google_project" "test_project" {
	project_id  = "%{project_name}"
}

data "google_kms_key_ring" "test_key_ring" {
	name     = "%{key_ring}"
	location = "us-central1"
}

data "google_kms_crypto_key" "key1" {
	name     = "%{key1}"
	key_ring = data.google_kms_key_ring.test_key_ring.id
}

  
resource "google_kms_crypto_key_iam_member" "key1_member" {
	crypto_key_id = data.google_kms_crypto_key.key1.id
	role      = "roles/cloudkms.cryptoKeyEncrypterDecrypter"

	member = "serviceAccount:service-${data.google_project.test_project.number}@gcp-sa-eventarc.iam.gserviceaccount.com"
}

resource "google_eventarc_channel" "primary" {
	location = "%{region}"
	name     = "tf-test-name%{random_suffix}"
	crypto_key_name =  data.google_kms_crypto_key.key1.id
	third_party_provider = "projects/${data.google_project.test_project.project_id}/locations/%{region}/providers/datadog"
	depends_on = [google_kms_crypto_key_iam_member.key1_member]
}
`, context)
}

func testAccEventarcChannel_cryptoKeyUpdate(context map[string]interface{}) string {
	return acctest.Nprintf(`
data "google_project" "test_project" {
	project_id  = "%{project_name}"
}

data "google_kms_key_ring" "test_key_ring" {
	name     = "%{key_ring}"
	location = "us-central1"
}
	
data "google_kms_crypto_key" "key2" {
	name     = "%{key2}"
	key_ring = data.google_kms_key_ring.test_key_ring.id
}

resource "google_kms_crypto_key_iam_member" "key2_member" {
	crypto_key_id = data.google_kms_crypto_key.key2.id
	role      = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
	
	member = "serviceAccount:service-${data.google_project.test_project.number}@gcp-sa-eventarc.iam.gserviceaccount.com"
}

resource "google_eventarc_channel" "primary" {
	location = "%{region}"
	name     = "tf-test-name%{random_suffix}"
	crypto_key_name= data.google_kms_crypto_key.key2.id
	third_party_provider = "projects/${data.google_project.test_project.project_id}/locations/%{region}/providers/datadog"
	depends_on = [google_kms_crypto_key_iam_member.key2_member]
}
`, context)
}