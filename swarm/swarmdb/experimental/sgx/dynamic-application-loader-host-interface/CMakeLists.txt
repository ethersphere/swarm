#cmake_minimum_required(VERSION 3.4)
cmake_minimum_required(VERSION 2.8.12)
project(jhi)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11 -Wall")

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin_linux)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin_linux)
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin_linux)

####################
#   teetransport   #
####################

set(TEETRANSPORT_HEADER_FILES
    common/include/teetransport.h
    teetransport/teetransport_internal.h)

set(TEETRANSPORT_SOURCE_FILES
    teetransport/teetransport.c
    teetransport/teetransport_internal.c
    teetransport/transport/socket/teetransport_socket.c
    teetransport/transport/socket/teetransport_socket_wrapper.c
    teetransport/transport/socket/lib/socket_linux.c
    teetransport/transport/libtee/teetransport_libtee.c
    teetransport/transport/libtee/teetransport_libtee_wrapper.c
    teetransport/transport/libtee/teetransport_libtee_client_metadata.c
    teetransport/transport/dal_device/teetransport_dal_device.c
    teetransport/transport/dal_device/teetransport_dal_device_wrapper.c
    thirdparty/libtee/linux/libteelinux.c
    thirdparty/libtee/linux/libmei/mei.c)

add_library(teetransport SHARED ${TEETRANSPORT_SOURCE_FILES} ${TEETRANSPORT_HEADER_FILES})

target_include_directories(teetransport PUBLIC
    teetransport/transport/socket
    teetransport/transport/socket/lib
    teetransport/transport/libtee
    teetransport/transport/dal_device
    teetransport
    common/include
    thirdparty/libtee/include/libtee
    thirdparty/libtee/linux/libmei)

#################
#   bhplugin1   #
#################

set(BHPLUGIN1_HEADER_FILES
    plugins/bhplugin1/jhi_plugin.h
    plugins/bhplugin1/jhi_plugin_types.h)

set(BHPLUGIN1_SOURCE_FILES
    plugins/bhplugin1/jhi_plugin.cpp
    thirdparty/bhplugin1/Beihai/tools/jhi_lib/BeihaiPlugin.cpp)

add_library(bhplugin1 SHARED ${BHPLUGIN1_SOURCE_FILES} ${BHPLUGIN1_HEADER_FILES})

target_include_directories(bhplugin1 PUBLIC
    thirdparty/bhplugin1/Beihai/tools/jhi_lib
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/include)

#################
#   bhplugin2   #
#################

set(BHPLUGIN2_HEADER_FILES
    plugins/bhplugin2/jhi_plugin.h
    plugins/bhplugin2/jhi_plugin_internal.h)

set(BHPLUGIN2_SOURCE_FILES
    plugins/bhplugin2/jhi_plugin.cpp
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl.cpp
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_platform_linux.cpp
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl_admin.cpp
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/impl/bhp_impl_ta.cpp
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/bh_acp_util.c
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/admin_pack_ext.c
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/admin_pack/admin_pack_int.c)

add_library(bhplugin2 SHARED ${BHPLUGIN2_SOURCE_FILES} ${BHPLUGIN2_HEADER_FILES})

target_include_directories(bhplugin2 PUBLIC
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/bhp/include
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/Beihai/shared/include
    thirdparty/bhplugin2/FW/src/apps/dal_ivm/beihai_shared/include
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/include)

##############
#   libjhi   #
##############

set(LIBJHI_HEADER_FILES common/include/jhi.h)

set(LIBJHI_SOURCE_FILES
    libjhi/jhi.cpp
    libjhi/CommandInvoker.cpp
    libjhi/CommandsClientSocketsLinux.cpp
    common/dbg-linux.cpp
    common/dbg.cpp
    common/locker-pthread.cpp
    common/jhi_event_linux.cpp
    common/reg-linux.cpp
    common/misc.cpp)

add_library(jhi SHARED ${LIBJHI_SOURCE_FILES} ${LIBJHI_HEADER_FILES})

target_include_directories(jhi PUBLIC
    libjhi
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/include)

target_link_libraries(jhi uuid pthread)

#####################
#   teemanagement   #
#####################

set(TEEMANAGEMENT_HEADER_FILES )

set(TEEMANAGEMENT_SOURCE_FILES
    teemanagement/teemanagement.cpp
    common/locker-pthread.cpp)

add_library(teemanagement SHARED ${TEEMANAGEMENT_SOURCE_FILES} ${TEEMANAGEMENT_HEADER_FILES})

target_include_directories(teemanagement PUBLIC
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    libjhi
    common/include)

target_link_libraries(teemanagement jhi pthread)

############
#   jhid   #
############

set(JHID_HEADER_FILES )

set(JHID_SOURCE_FILES
    service/JHIMain.cpp
    service/GlobalsManager.cpp
    service/ReadWriteLockPThread.cpp
    service/init.cpp
    service/EventManager.cpp
    service/jhi_plugin_loader.cpp
    service/AppletsManager.cpp
    service/SessionsManager.cpp
    service/install.cpp
    service/DLL_Loader.cpp
    service/AppletsPackageReader.cpp
    service/uninstall.cpp
    service/appProp.cpp
    service/FWInfoLinux.cpp
    service/closeSession.cpp
    service/XmlReaderLibXml2.cpp
    service/createSession.cpp
    service/LinuxService.cpp
    service/CommandsServerSocketsLinux.cpp
    service/CommandDispatcher.cpp
    service/getSCount.cpp
    service/getSessionStat.cpp
    service/sar.cpp
    service/SendCmdPkg.cpp
    external/libb64-1.2/src/cdecode.c
    common/dbg-linux.cpp
    common/dbg.cpp
    common/locker-pthread.cpp
    common/reg-linux.cpp
    common/misc.cpp
    common/jhi_event_linux.cpp
    common/jhi_semaphore-linux.cpp)

add_executable(jhid ${JHID_SOURCE_FILES} ${JHID_HEADER_FILES})

include(FindLibXml2)

target_include_directories(jhid PUBLIC
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/FWUpdate
    common/include
    external/libb64-1.2/include
    ${LIBXML2_INCLUDE_DIR})

target_link_libraries(jhid pthread dl teetransport uuid xml2)

#################
#   SmokeTest   #
#################

set(SMOKETEST_HEADER_FILES )

set(SMOKETEST_SOURCE_FILES
    test/smoketest/smoketest.cpp
    common/reg-linux.cpp
    common/misc.cpp)

add_executable(smoketest ${SMOKETEST_SOURCE_FILES} ${SMOKETEST_HEADER_FILES})

target_include_directories(smoketest PUBLIC
    test/smoketest
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/include)

target_link_libraries(smoketest jhi teemanagement pthread)

############
#   BIST   #
############

set (BIST_SOURCE_FILES test/bist/bist.cpp)

add_executable(bist ${BIST_SOURCE_FILES})

target_include_directories(bist PUBLIC
    thirdparty/bhplugin2/FW/src/apps/dal_ivm
    common/include)

target_link_libraries(bist jhi teemanagement)

###############
#   Install   #
###############

set(BINDIR       /usr/sbin)
set(LIBDIR       /usr/lib)
set(APPLETS_DIR  /var/lib/intel/dal/applets)
set(APP_REPO_DIR /var/lib/intel/dal/applet_repository)
set(CONFDIR      /etc/jhi)

# Choose init system (define init script install location)
if(${INIT_SYSTEM} MATCHES SysVinit)
    set(INIT_DIR /etc/init.d)
else() # systemd, default.
    execute_process(COMMAND pkg-config systemd --variable=systemdsystemunitdir
                    OUTPUT_VARIABLE SYSTEMD_DIR
                    OUTPUT_STRIP_TRAILING_WHITESPACE)
    #Enable when moving to CMake 3.4 instead of the current variable definition
    #pkg_get_variable(SYSTEMD_DIR systemd systemdsystemunitdir)
endif()

install(DIRECTORY DESTINATION ${APPLETS_DIR})
install(DIRECTORY DESTINATION ${APP_REPO_DIR})
install(DIRECTORY DESTINATION ${CONFDIR})

install(TARGETS
    jhid
    DESTINATION ${BINDIR})

install(TARGETS
    bhplugin1
    bhplugin2
    jhi
    teemanagement
    teetransport
    DESTINATION ${LIBDIR})

install(FILES
    applets/SpoolerApplet.dalp
    DESTINATION ${APPLETS_DIR})

install(FILES
    service/linux/jhi.conf
    DESTINATION ${CONFDIR})

# Choose init system (choose init script)
if(${INIT_SYSTEM} MATCHES SysVinit)
    install(PROGRAMS
        service/linux/jhi
        DESTINATION ${INIT_DIR})
else() # systemd, default.
    install(FILES
        service/linux/jhi.service
        DESTINATION ${SYSTEMD_DIR})
endif()