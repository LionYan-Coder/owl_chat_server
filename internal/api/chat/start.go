package chat

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	chatmw "github.com/openimsdk/chat/internal/api/mw"
	"github.com/openimsdk/chat/internal/api/util"
	"github.com/openimsdk/chat/pkg/common/config"
	"github.com/openimsdk/chat/pkg/common/imapi"
	"github.com/openimsdk/chat/pkg/common/kdisc"
	adminclient "github.com/openimsdk/chat/pkg/protocol/admin"
	chatclient "github.com/openimsdk/chat/pkg/protocol/chat"
	"github.com/openimsdk/tools/errs"
	"github.com/openimsdk/tools/mw"
	"github.com/openimsdk/tools/utils/datautil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	ApiConfig config.API
	Discovery config.Discovery
	Share     config.Share
}

func Start(ctx context.Context, index int, config *Config) error {
	if len(config.Share.ChatAdmin) == 0 {
		return errs.New("share chat admin not configured")
	}
	apiPort, err := datautil.GetElemByIndex(config.ApiConfig.Api.Ports, index)
	if err != nil {
		return err
	}
	client, err := kdisc.NewDiscoveryRegister(&config.Discovery)
	if err != nil {
		return err
	}

	chatConn, err := client.GetConn(ctx, config.Share.RpcRegisterName.Chat, grpc.WithTransportCredentials(insecure.NewCredentials()), mw.GrpcClient())
	if err != nil {
		return err
	}
	adminConn, err := client.GetConn(ctx, config.Share.RpcRegisterName.Admin, grpc.WithTransportCredentials(insecure.NewCredentials()), mw.GrpcClient())
	if err != nil {
		return err
	}
	chatClient := chatclient.NewChatClient(chatConn)
	adminClient := adminclient.NewAdminClient(adminConn)
	im := imapi.New(config.Share.OpenIM.ApiURL, config.Share.OpenIM.Secret, config.Share.OpenIM.AdminUserID)
	base := util.Api{
		ImUserID:        config.Share.OpenIM.AdminUserID,
		ProxyHeader:     config.Share.ProxyHeader,
		ChatAdminUserID: config.Share.ChatAdmin[0],
	}
	adminApi := New(chatClient, adminClient, im, &base)
	mwApi := chatmw.New(adminClient)
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery(), mw.CorsHandler(), mw.GinParseOperationID())
	SetChatRoute(engine, adminApi, mwApi)
	return engine.Run(fmt.Sprintf(":%d", apiPort))
}

func SetChatRoute(router gin.IRouter, chat *Api, mw *chatmw.MW) {

	config := router.Group("/config")
	config.POST("/checkVersion", chat.CheckVersion)
	config.POST("/fakeUser", chat.GetFakeUser)

	account := router.Group("/account")
	account.POST("/register", mw.CheckAdminOrNil, chat.RegisterUser) // Register
	account.POST("/login", chat.Login)                               // Login

	post := router.Group("/post", mw.CheckToken)
	post.POST("/publish", chat.PublishPost)
	post.POST("/like", chat.ChangeLikePost)
	post.POST("/collect", chat.ChangeCollectPost)
	post.POST("/forward", chat.ForwardPost)
	post.POST("/comment", chat.CommentPost)
	post.POST("/pin", chat.PinPost)
	post.POST("/reference", chat.ReferencePost)
	post.POST("/delete", chat.DeletePost)
	post.POST("/:postID", chat.GetPostByID)
	post.POST("/list_by_user", chat.GetPostListByUser)
	post.POST("/list", chat.GetPostList)
	post.POST("/list_all_type", chat.GetAllTypePost)
	post.POST("/comment_list", chat.GetCommentPostListByPostID)
	post.POST("/change_allow_comment", chat.ChangeAllowCommentPost)
	post.POST("/change_allow_forward", chat.ChangeAllowForwardPost)

	user := router.Group("/user", mw.CheckToken)
	user.POST("/update", chat.UpdateUserInfo)              // Edit personal information
	user.POST("/find/public", chat.FindUserPublicInfo)     // Get user's public information
	user.POST("/find/full", chat.FindUserFullInfo)         // Get all information of the user
	user.POST("/search/full", chat.SearchUserFullInfo)     // Search user's public information
	user.POST("/search/public", chat.SearchUserPublicInfo) // Search all information of the user
	user.POST("/search", chat.FindUserByAddressOrAccount)
	user.POST("/rtc/get_token", chat.GetTokenForVideoMeeting) // Get token for video meeting for the user
	user.POST("/statistic", chat.GetStatistic)
	user.POST("/online_time", chat.GetUsersOnlineTime)

	group := router.Group("/group", mw.CheckToken)
	group.POST("/contact/get", chat.GetGroupFromContact)
	group.POST("/contact/save", chat.SaveGroupToContact)
	group.POST("/contact/delete", chat.DeleteGroupFromContact)

	router.POST("/friend/search", mw.CheckToken, chat.SearchFriend)
	router.POST("/callback/:command", chat.OpenIMCallback)

	router.Group("/applet").POST("/find", mw.CheckToken, chat.FindApplet) // Applet list

	router.Group("/client_config").POST("/get", chat.GetClientConfig) // Get client initialization configuration
}
