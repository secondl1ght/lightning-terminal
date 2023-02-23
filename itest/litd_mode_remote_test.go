package itest

import (
	"context"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/lightninglabs/lightning-terminal/litrpc"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/stretchr/testify/require"
)

// testModeRemote makes sure that in remote mode all daemons work correctly.
func testModeRemote(ctx context.Context, net *NetworkHarness, t *harnessTest) {
	// Some very basic functionality tests to make sure lnd is working fine
	// in remote mode.
	net.SendCoins(t.t, btcutil.SatoshiPerBitcoin, net.Bob)

	// We expect a non-empty alias (truncated node ID) to be returned.
	resp, err := net.Bob.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	require.NoError(t.t, err)
	require.NotEmpty(t.t, resp.Alias)
	require.Contains(t.t, resp.Alias, "0")

	t.t.Run("certificate check", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		// In remote mode we expect the LiT HTTPS port (8443 by default)
		// and to have its own certificate
		litCerts, err := getServerCertificates(cfg.LitAddr())
		require.NoError(tt, err)
		require.Len(tt, litCerts, 1)
		require.Equal(
			tt, "litd autogenerated cert",
			litCerts[0].Issuer.Organization[0],
		)
	})
	t.t.Run("gRPC macaroon auth check", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runGRPCAuthTest(
					ttt, cfg.LitAddr(), cfg.LitTLSCertPath,
					endpoint.macaroonFn(cfg),
					endpoint.requestFn,
					endpoint.successPattern,
				)
			})
		}
	})

	t.t.Run("UI password auth check", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runUIPasswordCheck(
					ttt, cfg.LitAddr(), cfg.LitTLSCertPath,
					cfg.UIPassword, endpoint.requestFn,
					false, endpoint.successPattern,
				)
			})
		}
	})

	t.t.Run("UI index page fallback", func(tt *testing.T) {
		runIndexPageCheck(tt, net.Bob.Cfg.LitAddr())
	})

	t.t.Run("grpc-web auth", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runGRPCWebAuthTest(
					ttt, cfg.LitAddr(), cfg.UIPassword,
					endpoint.grpcWebURI,
				)
			})
		}
	})

	t.t.Run("gRPC super macaroon auth check", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		superMacFile, err := bakeSuperMacaroon(cfg, true)
		require.NoError(tt, err)

		defer func() {
			_ = os.Remove(superMacFile)
		}()

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runGRPCAuthTest(
					ttt, cfg.LitAddr(), cfg.LitTLSCertPath,
					superMacFile,
					endpoint.requestFn,
					endpoint.successPattern,
				)
			})
		}
	})

	t.t.Run("REST auth", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		for _, endpoint := range endpoints {
			endpoint := endpoint

			if endpoint.restWebURI == "" {
				continue
			}

			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runRESTAuthTest(
					ttt, cfg.LitAddr(), cfg.UIPassword,
					endpoint.macaroonFn(cfg),
					endpoint.restWebURI,
					endpoint.successPattern,
					endpoint.restPOST,
				)
			})
		}
	})

	t.t.Run("lnc auth", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		ctx := context.Background()
		ctxt, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()

		rawLNCConn := setUpLNCConn(
			ctxt, tt, cfg.LitAddr(), cfg.LitTLSCertPath,
			cfg.LitMacPath,
			litrpc.SessionType_TYPE_MACAROON_READONLY, nil,
		)
		defer rawLNCConn.Close()

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				runLNCAuthTest(
					ttt, rawLNCConn, endpoint.requestFn,
					endpoint.successPattern,
					endpoint.allowedThroughLNC,
					"unknown service",
				)
			})
		}
	})

	t.t.Run("lnc auth custom mac perms", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		ctx := context.Background()
		ctxt, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()

		customPerms := make(
			[]*litrpc.MacaroonPermission, 0, len(customURIs),
		)

		customURIKeyword := macaroons.PermissionEntityCustomURI
		for uri := range customURIs {
			customPerms = append(
				customPerms, &litrpc.MacaroonPermission{
					Entity: customURIKeyword,
					Action: uri,
				},
			)
		}

		rawLNCConn := setUpLNCConn(
			ctxt, tt, cfg.LitAddr(), cfg.LitTLSCertPath,
			cfg.LitMacPath,
			litrpc.SessionType_TYPE_MACAROON_CUSTOM, customPerms,
		)
		defer rawLNCConn.Close()

		for _, endpoint := range endpoints {
			endpoint := endpoint
			tt.Run(endpoint.name+" lit port", func(ttt *testing.T) {
				allowed := customURIs[endpoint.grpcWebURI]
				runLNCAuthTest(
					ttt, rawLNCConn, endpoint.requestFn,
					endpoint.successPattern,
					allowed, "permission denied",
				)
			})
		}
	})

	t.t.Run("gRPC super macaroon account system test", func(tt *testing.T) {
		cfg := net.Bob.Cfg

		superMacFile, err := bakeSuperMacaroon(cfg, false)
		require.NoError(tt, err)

		defer func() {
			_ = os.Remove(superMacFile)
		}()

		ht := newHarnessTest(tt, net)
		runAccountSystemTest(
			ht, net.Bob, cfg.LitAddr(), cfg.LitTLSCertPath,
			superMacFile, 1,
		)
	})
}
