#ifndef __OPAPI2_H
#define __OPAPI2_H

#if defined(__WIN32) || defined(__WIN32__) || defined(WIN32)
    #ifndef MAGUS_WINDOWS
        #define MAGUS_WINDOWS
    #endif
#else
    #ifndef MAGUS_POSIX
        #define MAGUS_POSIX
    #endif
#endif

#ifndef GMacroAPI
    #ifdef MAGUS_WINDOWS
        #ifdef MAGUS_IMPLETMENT_SHARED
            #define GMacroAPI(type) __declspec(dllexport) type _stdcall
        #elif defined MAGUS_IMPLETMENT_STATIC || defined MAGUS_USE_STATIC
            #define GMacroAPI(type) type
        #else
            #define GMacroAPI(type) __declspec(dllimport) type _stdcall
        #endif
    #else
        #define GMacroAPI(type) type
    #endif
#endif

/* ============================================================================================================================ */
//说明
/**
 *  @mainpage openPlant apiV2
 * 
 *  使用openPlant apiV2 的一般步骤：
 *
 *      首先链接openPlant,使用op2_init,保存其返回的句柄
 *          如果成功，则返回一个句柄，
 *          如果失败，则返回NULL。
 *
 *      然后使用该句柄，传递给相应API执行数据请求。
 *      处理API的返回码，检查errors返回值。在保存取到的数据，若数据为一个结果集，则再调用相应API对结果集进一步处理。
 *
 *      若有对象需要释放内存，则释放该对象内存。
 *
 *      关闭openPlant句柄。
 *
 */

/* ============================================================================================================================ */
/**
 * @defgroup openplant_c_api2 openPlant C APIv2
 * @brief openPlant C APIv2，一组openPlant读写的C结构和函数
 * @{
 */
#ifdef __cplusplus
extern "C"
{
#endif

/* ============================================================================================================================ */
/**
 * @defgroup api2_marco APIv2 常量定义
 * @brief APIv2 使用的一组常量
 * @{
 */
// 点的记录类型
#define AX_TYPE   0  ///< 模拟类型
#define DX_TYPE   1  ///< 数字类型
#define I2_TYPE   2  ///< 组合类型
#define I4_TYPE   3  ///< 整数类型
#define R8_TYPE   4  ///< 双精度类型

// 点的数据来源
#define IN_POINT  0 ///< 采集点
#define CA_POINT  1 ///< 计算点

#define OP_TIMEOUT   0x8000 ///< 点值的超时，如果设置为1，则为超时
#define OP_BAD       0x0300 ///< 点值的好坏，如果设置为1，则为坏值

// 历史类型及常量
#define OPR_HIS_RAW      0    ///< 取原始值
#define OPR_HIS_SPAN     1    ///< 取等间距值
#define OPR_HIS_PLOT     2    ///< 取PLOT值，每区间包括（起始值，最大值，最小值, 结束）
#define OPR_HIS_FLOW     8    ///< 取累计值
#define OPR_HIS_MAX      9    ///< 取最大值
#define OPR_HIS_MIN      10   ///< 取最小值
#define OPR_HIS_AVG      11   ///< 取平均值
#define OPR_HIS_MEAN     12   ///< 取算术平均值
#define OPR_HIS_STDEV    13   ///< 取标准方差值，未实现
#define OPR_HIS_SUM      14   ///< 取原始值的累加和，
#define OPR_HIS_SUMMARY  15   ///< 取所有统计值：累计/最大/最小/平均/方差

// 服务器选项
#define OP_OPTION_WALL_0BIT    1 ///< 客户端与服务器中间经过单向隔离器
#define OP_OPTION_WALL_1BIT    2 ///< 客户端与服务器中间经过单向隔离器
#define OP_OPTION_BUFFER       8 ///< 在当前句柄启用失败后本地缓存
#define OP_OPTION_DEBUG_INFO   256 ///< 产生出错信息，帮助诊断问题
#define OP_OPTION_LOG          512 ///< 产生出错信息，帮助诊断问题

// 数据库对象
#define OP_DATABASE            0x01 ///<根节点对象
#define OP_NODE                0x10 ///<节点对象
#define OP_AX_POINT            0x20 ///<模拟量点对象
#define OP_DX_POINT            0x21 ///<数字量点对象
#define OP_I2_POINT            0x22 ///<短整数点对象
#define OP_I4_POINT            0x23 ///<长整数点对象
#define OP_R8_POINT            0x24 ///<双精度点对象

// 数据库操作
#define OP_SELECT             0 ///< 查询现有值
#define OP_UPDATE             1 ///< 更新现有值
#define OP_INSERT             2 ///< 接入新值
#define OP_DELETE             3 ///< 删除现有值
#define OP_REPLACE            4 ///< 如果有就更新，如没有就插入

// 错误代码
#define OPE_OK             0     ///< 成功，没有错误
#define OPE_ERROR         -1     ///< 未知错误
#define OPE_PARAM         -10    ///< 参数错误
#define OPE_UNSUPPORTED   -11    ///< 功能未支持
#define OPE_MEMORY        -96    ///< 无法分配内存，需要重连
#define OPE_NET_IO        -97    ///< 网络读写IO错误，需要重连
#define OPE_NET_CLOSED    -98    ///< 连接已关闭，需要重连
#define OPE_CONNECT       -99    ///< 无法连接服务器，需要重连
#define OPE_BUFF_NET      -1001  ///< 网络断开
#define OPE_BUFF_IO       -1002  ///< 写入存储缓存文件出错了，最有可能是文件所在分区满了
#define OPE_BUFF_OVERFLOW -1003  ///< 数据文件到了预设的大小
/** @} */


typedef int OPDateTime_t;

/* ============================================================================================================================ */
/**
 * @defgroup api2_type_define APIv2 基本类型定义
 * @brief 定义一些最基础的常量与类型
 * @{
 */

/**
* @brief 统计值
*/
typedef struct
{
    int time;        // 时间标签，统计区间的起始时间
    short status;       // 统计区间内记录质量的百分比0-100
    double flow;        // 累积
    double max;         // 最大，DX: 跳变为1的次数
    double min;         // 最小，DX: 跳变为0的次数
    int maxtime;       // 最大值时间，DX: 值为1持续时间
    int mintime;       // 最小值时间，DX: 值为0持续时间
    double avg;         // 时间平均
    double mean;        // 算术平均
    double stdev;       // 标准方差, 尚未实现
}StatVal;
/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup api2_handle_face APIv2 有连接接口
 * @brief 把无连接部分接口抽取出来，做成有连接接口；它们的功能都一对应的
 *
 * @details
 *  @li 这一组函数从 @ref api2_nohandle_face 里抽取出来然后加入连接内容，
 *          其工作实际内容是相同的
 *  @li 在 @ref api2_nohandle_face 里的UDP函数和静态数据写函数没有抽取出来
 *  @li @ref api2_nohandle_face 是有连接接口的超集，但有连接接口提供一组线程安全的函数，
 *          同时重新统一设计了函数命名，使得对函数的功能更容易从函数名称上体现
 *
 * @{
 */

/**
* @brief 数据库服务器，用户透明
*/
typedef void* OpenPlant;
/**
* @brief 结果集，用户透明
*/
typedef void* OPResult;
/**
* @brief 本地缓冲，用户透明
*/
typedef void* OPLocalBuffer;

/**
* @brief openPlant 客户端 API v2 初始化
*
* @param option   选项; 参考宏: OP_OPTION_*
* @param host     服务器IP或名字
* @param port     服务器端口
* @param timeout  网络超时，以秒为单位
* @param user     用户名
* @param password 密码
* @buffer_path    缓存所在目录；如果不需要缓存指定为 nullptr
* @buffer_size    缓存文件尺寸；如果不需要缓存指定为 0
*
* @return
*   @li 成功: ret != 0，这是一个句柄地址
*   @li 失败: ret == 0
*/
GMacroAPI(OpenPlant) op2_init(int option
        , const char *host , int port , int timeout
        , const char *user , const char *password
        , const char *buffer_path , int buffer_size);

/**
* @brief 关闭连接
*
* @param fd 指定连接
* @return 本函数没有返回值
*/
GMacroAPI(void) op2_close(OpenPlant op);

/**
* @brief 查询服务器连接状态
*
* @param op 数据库服务器句柄
* @return 0 表示OK，-1 表示已关闭
*/
GMacroAPI(int) op2_status(OpenPlant op);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup apiv2_op_buffer opWriter 属性设置接口
 * @brief   对OPBuffer属性值进行设置
 * @details
 *      key、type、意义、[默认值]
 *  @li transfer_protocol                int     使用tcp传输         无默认值
 *  @li server_address                   str     服务器地址          无默认值
 *  @li server_port                      int     服务器端口号        无默认值
 *  @li client_name                      str     客户端用户名        无默认值
 *  @li client_password                  str     客户端密码          无默认值
 *  @li storage_location                 str     缓存数据存储路径    无默认值
 *  @li storage_capacity                 int     缓存数据最大容量    无默认值
 *  @li option                           int     见op_writer_open    无默认值
 *  @li isolator_enabled                 int     过隔离器            无默认值
 *  @li net_probe_interval               int     网络探测间隔        5 second
 *  @li upload_history_interval          int     上传历史间隔        100 ms
 *  @li realtime_filter_enabled          int     是否过滤实时值      0 bool
 *  @li filter_bool_upload_interval      int     开关量上传间隔      30 second
 *  @li filter_bool_storage_interval     int     开关量存储间隔      900 second
 *  @li size_for_each_upload             int     每次上传文件大小    10 MB
 *  @li upload_history_per_time          int     每次上传的记录数    50000 条
 *  @li single_file_capacity             int     每个缓存文件大小    1024 MB
 *  @li storage_file_quantity            int     缓冲的最大文件个数  无默认值
 *  @li storage_time_limit               int     缓冲的最大天数      无默认值
 *  @li history_task_interval            int     后台线程sleep时间   1000 ms
 *  @li record_filename                  str     记录文件名          "op.writer.json"
 *
 * @{
 */


/**
* @brief 获取OPBuffer的内部状态
*
* @param op 数据库句柄
* @param key 关键字(key):  network_connection, try_connection_count, cache_status, upload_status
* @param value 对应属性的值
*/
GMacroAPI(int) op2_buffer_get_status(OpenPlant op, const char* key);

/**
* @brief 对OPBuffer属性值进行设置，其属性值类型为int
*
* @param op 数据库句柄
* @param name 属性关键字(key),如:transfer_protocol (使用tcp传输)
* @param value 对应属性的值
*/
GMacroAPI(void) op2_buffer_set_int   (OpenPlant op, const char* name, int value);

/**
* @brief 对OPBuffer属性值进行设置，其属性值类型为string
*
* @param
* @param name 属性关键字(key)
* @param value 对应属性的值

*/
GMacroAPI(void) op2_buffer_set_string(OpenPlant op, const char* name, const char* value);

/**
* @brief 对点进行设置
* @param op 数据库句柄
* @param id  点的ID
* @param type 点的类型
* @param deadband 死区值
*/
GMacroAPI(void) op2_buffer_set_point (OpenPlant op, int id, int type, double deadband);


/** 
 * @brief 打开本地缓冲，按照给定时间段从中检索数据
 * 
 * @param directroy     缓冲文件所在目录
 * @param beginTime     要检索数据的开始时间
 * @param endTime       要检索数据的结束时间
 * @param error         错误字符串
 *
 * @return 成功:OPLocalBuffer; 失败:NULL
 */
GMacroAPI(OPLocalBuffer) op2_buffer_open(const char* directroy, OPDateTime_t beginTime, OPDateTime_t endTime, const char** error);

/** 
 * @brief 获取缓冲中的下一块
 * 
 * @param opBuffer op2_buffer_open返回的句柄
 *
 * @return 有数据:返回一个数据块(结果集); 无数据:NULL
 */
GMacroAPI(OPResult) op2_buffer_next(OPLocalBuffer opBuffer);

/** 
 * @brief 
 * 
 * @param result    op2_buffer_next 返回的数据块
 * @param id        数据块中下一个记录中的id字段
 * @param time      数据块中下一个记录中的time字段
 * @param status    数据块中下一个记录中的status字段
 * @param value     数据块中下一个记录中的value字段
 */
GMacroAPI(int) op2_buffer_get_from_result(OPResult result, int *id, int *time, short *status, double *value);

/** 
 * @brief 释放op2_buffer_next 返回的结果集
 * 
 * @param 将被释放的结果集
 */
GMacroAPI(void) op2_buffer_free_result(OPResult result);

/** 
 * @brief 关闭op2_buffer_open打开的本地缓冲
 * 
 * @param opBuffer 将被关闭的本地缓冲
 */
GMacroAPI(void) op2_buffer_close(OPLocalBuffer opBuffer);


/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup api2_time_face APIv2 时间操作函数
 * @brief 把标量时间与分段时间互换
 *
 * @{
 */

/**
* @brief 取得数据库服务器时间
*
* @param op 数据库服务器句柄
* @param out 返回数据库时间(std::time_t 时间值)
* @return 正确 返回 0
*         错误 返回 其他

*/
GMacroAPI(int) op2_get_system_time(OpenPlant op, int*out);

/**
* @brief 把 std::time_t 时间值转换为 分段时间值，即 2011-11-23 14:14:20
*
* @param t time_t 类型时间
* @param yy 返回的年
* @param mm 返回的月
* @param dd 返回的日
* @param hh 返回的时
* @param mi 返回的分
* @param ss 返回的秒
*
* @return 本函数不返回值
*/
GMacroAPI(void) op2_decode_time(int time, int *yy, int *mm, int *dd, int *hh, int *mi, int *ss);

/**
* @brief 把 分段时间值，即 2011-11-23 14:14:20，转换为  std::time_t 时间值
*
* @param yy 输入的年
* @param mm 输入的月
* @param dd 输入的日
* @param hh 输入的时
* @param mi 输入的分
* @param ss 输入的秒
*
* @return time_t 类型时间
*/
GMacroAPI(int) op2_encode_time(int yy, int mm, int dd, int hh, int mi, int ss);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup point_group APIv2 点组操作函数
 * @brief 点组操作
 * @{
 */

/**
* @brief 点组句柄，一组点名的代表
* @details
*   @li 有时我们需要同时操作多个点，并且此时只知道点名；那么点组就为这时准备的
*   @li 可以通过点组来完成“从点名到ID”映射，即通过名字取得ID
*/
typedef void* OPGroup;

/**
* @brief 创建指定点组
* @return 正确: 点组句柄；错误: NULL
*/
GMacroAPI(OPGroup) op2_new_group();

/**
* @brief 取到指定点组的大小
* @param gh 点组句柄，用来指定一个点组
*/
GMacroAPI(int) op2_group_size(OPGroup gh);

/**
* @brief 把指定点名加入点组
* @param gh  点组句柄，由 @ref op_new_group 返回
* @param obj_name 点名
*/
GMacroAPI(void) op2_add_group_point(OPGroup gh, const char *obj_name);

/**
* @brief 释放一个点组
* @param gh 点组句柄
*/
GMacroAPI(void) op2_free_group(OPGroup gh);

/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup api2_result_handle APIv2 结果集操作函数
 * @brief 结果集操作，取值或释放
 * @{
 */

/**
* @brief 取得结果集尺寸
* @param result 结果集
* @return 结果集尺寸
*/
GMacroAPI(int) op2_num_rows(OPResult result);

/**
* @brief 假设结果集是一个历史值结果集，取得其中下一个记录的值
*
* @param result 结果集
* @param value 用于返回历史值
* @param status 用于返回状态值
* @param time 用于返回时间标签
*
* @return
*   @li 0: 已经到结果集末尾
*   @li 1: 还有下一条记录
*   @li 其它: 出错
*/
GMacroAPI(int) op2_fetch_timed_value(OPResult result, int* time, short *status, double *value);

/**
* @brief 假设结果集是一个统计值结果集，取得其中下一个统计值
*
* @param result 结果集
* @param sval 用于返回的统计值
*
* @return
*   @li 0: 已经到结果集末尾
*   @li 1: 还有下一条记录
*   @li 其它: 出错
*/
GMacroAPI(int) op2_fetch_stat_value(OPResult result, StatVal *sval);

/**
* @brief 释放一个结果集
*
* @param result 结果集
* @return 本函数没有返回值
*/
GMacroAPI(void) op2_free_result(OPResult result);

/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup op2_get_face APIv2 取实时数据和历史数据相关接口
 * @brief   通过点名或者ID取实时数据或者是历史数据 
 *
 * @{
 */

/**
* @brief 通过点名取得实时值
*
* @param op 数据库句柄
* @param gh 点名的点组
* @param time 指向点的时间的数组
* @param status 指向点的状态的数组
* @param value  指向点的实时值的数组
* @param errors 指向回返错误码的数组，若数组中有为非零的值，则表明取该点的实时值错误
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_value_byname(OpenPlant op, OPGroup gh, int* time, short *status, double *value, int *errors);

/**
* @brief 通过ID取得实时值
*
* @param op  数据库句柄
* @param num 点ID的个数
* @param id  指向点ID的数组
* @param time 指向点的时间的数组
* @param status 指向点的状态的数组
* @param value  指向点的实时值的数组
* @param errors 指向回返错误码的数组，若数组中有为非零的值，则表明取该点的实时值错误
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_value_byid(OpenPlant op, int num, int *id, int* time, short *status, double *value, int *errors);

/**
* @brief 通过点名取得快照值
*
* @param op 数据库句柄
* @param gh 点名点组
* @param time 某一个时刻的值（std::time_t)
* @param status 指向点的状态的数组
* @param value  指向点对应某一时刻值的数组
* @param errors 指向回返错误码的数组，若数组中有为非零的值，则表明取该点的值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_snap_byname(OpenPlant op, OPGroup gh, int time, short *status, double *value, int *errors);

/**
* @brief 通过点ID取得快照值
* @param op 数据库句柄
* @param num 点ID的个数
* @param id  指向点ID的数组
* @param time 某一个时刻的值（std::time_t)
* @param status 指向点的状态的数组
* @param value  指向点对应某一时刻值的数组
* @param errors 指向回返错误码的数组，若数组中有为非零的值，则表明取该点的值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_snap_byid(OpenPlant op, int num, int *id, int time, short *status, double *value, int *errors);

/**
* @brief 通过点名取历史数据
*
* @param op 数据库句柄
* @param gh 点名点组
* @param value_type  历史数据类型, 如：原始值（0），等间距值（1）等等
* @param begin_tm    起始时间
* @param end_tm      结束时间
* @param interval    间隔时间
* @param result      历史结果集
* @param errors      指向回返错误码的数组，若数组中有为非零的值，则表明取该点的值错误
*
* @return 正确 返回 0
*        错误 返回 其他
*/
GMacroAPI(int) op2_get_history_byname(OpenPlant op, OPGroup gh, int *value_type, int begin_tm, int end_tm, int interval, OPResult *result, int *errors);

/**
* @brief  通过点ID取历史数据
*
* @param op 数据库句柄
* @param num 点ID个数
* @param id  指向点ID的数组
* @param value_type 历史数据类型, 如：原始值（0），等间距值（1）等等
* @param begin_tm   起始时间
* @param end_tm     结束时间
* @param interval   间隔时间
* @param result     历史结果集
* @param errors     指向回返错误码的数组，若数组中有为非零的值，则表明取该点的值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_history_byid(OpenPlant op, int num, int *id, int *value_type, int begin_tm, int end_tm, int interval, OPResult *result, int *errors);

/**
* @brief 把结果集中的原始值还原为等间距值
*
* @param result 历史结果集，由op2_get_history_byname或op2_get_history_byid得到
* @param num    用于返回数据条数 
* @param time   用于返回时间标签数组 
* @param status 用于返回状态数组 
* @param value  用于返回等间距值数组 
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_raw_to_span(OPResult result, int *num, int **time, short **status, double **value);

/**
* @brief 通过点名对一个点取一段历史数据，历史数据的条数有限制
*
* @param op         数据库服务器句柄
* @param obj_name    一个点名
* @param value_type 历史值类型
* @param begin_tm   起始时间
* @param end_tm     结束时间
* @param interval   时间间隔
* @param num        历史数据的条数不超过num
* @param time     指向时间标签数组
* @param status     指向状态数组
* @param value      指向值数组
* @param actSize    返回值实际的个数
*/
GMacroAPI(int) op2_get_histroy_top_byname(OpenPlant op, const char *obj_name, int value_type, int begin_tm, int end_tm, int interval,
                                      int num, int* time, short *status, double *value, int *actSize);

/**
* @brief 通过点id对一个点取一段历史数据，历史数据的条数有限制
*
* @param op         数据库服务器句柄
* @param id         一个点ID
* @param value_type 历史值类型
* @param begin_tm   起始时间
* @param end_tm     结束时间
* @param interval   时间间隔
* @param num        历史数据的条数不超过num
* @param time     指向时间标签数组
* @param status     指向状态数组
* @param value      指向值数组
* @param actSize    返回值实际的个数
*/
GMacroAPI(int) op2_get_histroy_top_byid(OpenPlant op, int id, int value_type, int begin_tm, int end_tm, int interval,
                                    int num, int* time, short *status, double *value, int *actSize);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup op2_write_face APIv2 写实时数据和历史数据相关接口
 * @brief   通过点名或者ID写实时数据或者是历史数据 
 *
 * @{
 */

/**
* @brief 向多点写某一时刻的实时值
*
* @param op      数据库句柄
* @param pt      点类型
* @param num     点个数
* @param id      指向点ID数组
* @param time  时刻值（std:time_t)
* @param status  指向点的状态的数组
* @param value   指向点的值的数组
* @param errors  指向回返错误码的数组，数组中值为0或者1都表示写入成功，其中1表示第一次写入成功；0表示对当前实时值修改成功。
*
* @return 正确 返回 0 
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_value(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief 向多点写某一时刻的实时值(与op2_write_value的区别是只向点写值而不写状态)
*
* @param op      数据库句柄
* @param pt      点类型
* @param num     点个数
* @param id      指向点ID数组
* @param time  时刻值（std:time_t)
* @param status  指向点的状态的数组
* @param value   指向点的值的数组
* @param errors  指向回返错误码的数组，数组中值为0或者1都表示写入成功，其中1表示第一次写入成功；0表示对当前实时值修改成功。
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_value_only(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief 向多点的实时值(与op2_write_value的区别是：每个点都对应一个时刻)
*
* @param op      数据库句柄
* @param pt      点类型
* @param num     点个数
* @param id      指向点ID数组
* @param time  指向时刻值的数组
* @param status  指向点的状态的数组
* @param value   指向点的值的数组
* @param errors  指向回返错误码的数组，数组中值为0或者1都表示写入成功，其中1表示第一次写入成功；0表示对当前实时值修改成功。
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_value_tm(OpenPlant op, int pt, int num, const int *id, const int* time, const short *status, const double *value, int *errors);

/**
* @brief 使用点名向一个点写多个历史数据
*
* @param op       数据库句柄
* @param pt       点的类型
* @param obj_name  一个点名
* @param num      写入数据的条数
* @param time   指向时间标签的数组
* @param status   指向对应与时间标签点状态的数组
* @param value    指向对于与时间标签点值的数组
* @param error   指向回返错误码的值，若该值非零，则表明该点写值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_histroy_byname(OpenPlant op, int pt, const char *obj_name, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief 使用点ID向一个点写多个历史数据
*
* @param op       数据库句柄
* @param pt       点的类型
* @param id       一个点的ID
* @param num      写入数据的条数
* @param time   指向时间标签的数组
* @param status   指向对应与时间标签点状态的数组
* @param value    指向对于与时间标签点值的数组
* @param error   指向回返错误码的值，若该值非零，则表明该点写值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_histroy_byid(OpenPlant op, int pt, int id, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief 使用ID向多个点在同一时刻写历史值
*
* @param op       数据库句柄
* @param pt       点类型
* @param num      点的个数
* @param id       指向点ID的数组
* @param time   时刻值
* @param status   指向点状态的数组
* @param value    指向点值的数组
* @param error   指向回返错误码的值，若该值非零，则表明该点写值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_snap(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief 向一个点写多条历史数据
*
* @param op      数据库句柄
* @param pt      点类型
* @param id      点ID
* @param num     写入历史数据条数
* @param time  指向点时间标签的数组
* @param status  指向对应与时间标签点状态的数组
* @param value   指向对于与时间标签点值的数组
* @param error   指向回返错误码的值，若该值非零，则表明该点写值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_cache_one(OpenPlant op, int pt, int id, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief 同时向多个点写历史数据
*
* @param op      数据库句柄
* @param pt      点类型
* @param num     点的个数
* @param id      指向点ID的数组
* @param time  指向点时间标签的数组
* @param status  指向对应与时间标签点状态的数组
* @param value   指向对于与时间标签点值的数组
* @param errors  指向回返错误码的数组，若数组中有为非零的值，则表明该点写实时值错误
*
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_write_cache(OpenPlant op, int pt, int num, const int *id, const int* time, const short *status, const double *value, int *errors);

/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup op2_object_handle APIv2 对数据库对象操作相关的接口
 * @brief  这组接口可以对数据库对象（节点，点）进行增删改查,比如加点，删点，取静态信息，修改静态信息等等
 *
 * @{
 */

/**
* @brief 数据对象，用户透明
*/
typedef void* OPObject;

/**
* @brief 通过点名取得对应ID
*
* @param op 数据库句柄
* @param gh 点名的点组
* @param id 指向点的ID的数组，用于存放返回点名对应的ID
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_get_id_byname(OpenPlant op, OPGroup gh, int *id);

/**
* @brief 创建对象
*
* @param op   数据库句柄
* @param name 要创建对象的名字
* @param type 指定对象类型（参考对象类型）
* @return
*   @li 成功: 对象描述字
*   @li 失败: NULL
*/
GMacroAPI(OPObject) op2_new_object(OpenPlant op, const char *name, int type);

/**
* @brief 释放对象内存空间
*
* @param o 对象句柄
*/
GMacroAPI(void) op2_free_object(OPObject o);

/**
* @brief 设置一个对象的静态属性值，其值类型为string
*
* @param o      对象描述字（由op2_new_object返回）
* @param field  静态属性的字段（如：PN、ED、EU 等等）
* @param value  静态属性字段的值
*/
GMacroAPI(void) op2_object_set_string(OPObject o, const char *field, char *value);

/**
* @brief 设置一个对象的静态属性值，其值类型为int
*
* @param o      对象描述字（由op2_new_object返回）
* @param field  静态属性的字段（如：PN、ED、EU 等等）
* @param value  静态属性字段的值
*/
GMacroAPI(void) op2_object_set_int   (OPObject o, const char *field, int value);

/**
* @brief 设置一个对象的静态属性值，其值类型为double
*
* @param o      对象描述字（由op2_new_object返回）
* @param field  静态属性的字段（如：PN、ED、EU 等等）
* @param value  静态属性字段的值
*/
GMacroAPI(void) op2_object_set_double(OPObject o, const char *field, double value);

/**
* @brief 取得一个对象的静态属性值，其值类型为string
*
* @param o       对象描述字（由op2_new_object返回）
* @param field   静态属性的字段（如：PN、ED、EU 等等）
* @param buf     字符数组，用来存放静态属性的值
* @param len     字符数组可用的最大长度
* @return 正确 返回 0
*         错误 返回 其他
*/
GMacroAPI(int) op2_object_get_string(OPObject o, const char *field, char *buf, int len);

/**
* @brief 取得一个对象的静态属性值，其值类型为int
*
* @param o       对象描述字（由op2_new_object返回）
* @param field   静态属性的字段（如：PN、ED、EU 等等）
* @return  返回 静态属性的值
*/
GMacroAPI(int) op2_object_get_int   (OPObject o, const char *field);

/**
* @brief 取得一个对象的静态属性值，其值类型为double
*
* @param o       对象描述字（由op2_new_object返回）
* @param field   静态属性的字段（如：PN、ED、EU 等等）
* @return  返回 静态属性的值
*/
GMacroAPI(double) op2_object_get_double(OPObject o, const char *field);

/**
* @brief 创建/删除/更新对象
*
* @param op      数据库句柄
* @param cmd     对数据库的操作 （OP_SELECT,OP_UPDATE,OP_INSERT,OP_DELETE,OP_REPLACE)
* @param parent  该组对象父节点的名字（全名）
* @param num     对象的个数
* @param objects 对象数组
* @param errors  错误码，若错误码中有0，则表示对该对象的操作未成功
*
* @return
*   @li 成功: 返回操作成功的对象个数
*   @li 失败: -1
*/
GMacroAPI(int) op2_modify_object(OpenPlant op, int cmd, const char *parent, int num, OPObject *objects, int *errors);

/**
* @brief 获取数据对象
*
* @param op 数据库服务器
* @param gh 名字组
* @param objects 对象数组
* @param errors 错误码
* @return
*   @li 成功: 0
*   @li 失败: -1
*/
GMacroAPI(int) op2_get_object_byname(OpenPlant op, OPGroup gh, OPObject *objects, int *errors);

/**
* @brief 获取数据对象
*
* @param op 数据库服务器
* @param num 对象个数
* @param num 对象ID数组
* @param objects 对象数组
* @param errors 错误码
* @return
*   @li 成功: 0
*   @li 失败: -1
*/
GMacroAPI(int) op2_get_object_byid(OpenPlant op, int num, int *id, OPObject *objects, int *errors);

/**
* @brief 获取数据库列表
*
* @param op 数据库服务器
* @param num          取到的数据库列表个数
* @param databases    获取到的数据库对象列表
* @return
*   @li 成功: 0
*   @li 失败: 错误码
*/
GMacroAPI(int) op2_get_database(OpenPlant op, int *num, OPObject **databases);

/**
* @brief 获取子对象列表
*
* @param op       数据库服务器句柄
* @param parent   指定父对象名称
* @param num      取到子对象个数
* @param objects  获取到的子对象列表
* @return
*   @li 成功: 0
*   @li 失败: 错误码
*/
GMacroAPI(int) op2_get_child(OpenPlant op, const char *parent, int *num, OPObject **objects);

/**
* @brief 获取子对象列表
*
* @param op     数据库服务器句柄
* @param parent 指定父对象名称
* @param num    取到子对象个数
* @param objects 获取到的子对象列表
* @return
*   @li 成功: 0
*   @li 失败: 错误码
*/
GMacroAPI(int) op2_get_child_idname(OpenPlant op, const char *parent, int *num, OPObject **objects);

/**
* @brief 释放对象列表
*
* @param num     对象列表的个数
* @param objects 对象列表
* @return 本函数没有返回值
*/
GMacroAPI(void) op2_free_list(int num, OPObject *objects);
/** @} */


#ifdef __cplusplus
}
#endif //__cplusplus
/** @} */

#endif //__OPAPI2_H


