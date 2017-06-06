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
//˵��
/**
 *  @mainpage openPlant apiV2
 * 
 *  ʹ��openPlant apiV2 ��һ�㲽�裺
 *
 *      ��������openPlant,ʹ��op2_init,�����䷵�صľ��
 *          ����ɹ����򷵻�һ�������
 *          ���ʧ�ܣ��򷵻�NULL��
 *
 *      Ȼ��ʹ�øþ�������ݸ���ӦAPIִ����������
 *      ����API�ķ����룬���errors����ֵ���ڱ���ȡ�������ݣ�������Ϊһ������������ٵ�����ӦAPI�Խ������һ������
 *
 *      ���ж�����Ҫ�ͷ��ڴ棬���ͷŸö����ڴ档
 *
 *      �ر�openPlant�����
 *
 */

/* ============================================================================================================================ */
/**
 * @defgroup openplant_c_api2 openPlant C APIv2
 * @brief openPlant C APIv2��һ��openPlant��д��C�ṹ�ͺ���
 * @{
 */
#ifdef __cplusplus
extern "C"
{
#endif

/* ============================================================================================================================ */
/**
 * @defgroup api2_marco APIv2 ��������
 * @brief APIv2 ʹ�õ�һ�鳣��
 * @{
 */
// ��ļ�¼����
#define AX_TYPE   0  ///< ģ������
#define DX_TYPE   1  ///< ��������
#define I2_TYPE   2  ///< �������
#define I4_TYPE   3  ///< ��������
#define R8_TYPE   4  ///< ˫��������

// ���������Դ
#define IN_POINT  0 ///< �ɼ���
#define CA_POINT  1 ///< �����

#define OP_TIMEOUT   0x8000 ///< ��ֵ�ĳ�ʱ���������Ϊ1����Ϊ��ʱ
#define OP_BAD       0x0300 ///< ��ֵ�ĺû����������Ϊ1����Ϊ��ֵ

// ��ʷ���ͼ�����
#define OPR_HIS_RAW      0    ///< ȡԭʼֵ
#define OPR_HIS_SPAN     1    ///< ȡ�ȼ��ֵ
#define OPR_HIS_PLOT     2    ///< ȡPLOTֵ��ÿ�����������ʼֵ�����ֵ����Сֵ, ������
#define OPR_HIS_FLOW     8    ///< ȡ�ۼ�ֵ
#define OPR_HIS_MAX      9    ///< ȡ���ֵ
#define OPR_HIS_MIN      10   ///< ȡ��Сֵ
#define OPR_HIS_AVG      11   ///< ȡƽ��ֵ
#define OPR_HIS_MEAN     12   ///< ȡ����ƽ��ֵ
#define OPR_HIS_STDEV    13   ///< ȡ��׼����ֵ��δʵ��
#define OPR_HIS_SUM      14   ///< ȡԭʼֵ���ۼӺͣ�
#define OPR_HIS_SUMMARY  15   ///< ȡ����ͳ��ֵ���ۼ�/���/��С/ƽ��/����

// ������ѡ��
#define OP_OPTION_WALL_0BIT    1 ///< �ͻ�����������м侭�����������
#define OP_OPTION_WALL_1BIT    2 ///< �ͻ�����������м侭�����������
#define OP_OPTION_BUFFER       8 ///< �ڵ�ǰ�������ʧ�ܺ󱾵ػ���
#define OP_OPTION_DEBUG_INFO   256 ///< ����������Ϣ�������������
#define OP_OPTION_LOG          512 ///< ����������Ϣ�������������

// ���ݿ����
#define OP_DATABASE            0x01 ///<���ڵ����
#define OP_NODE                0x10 ///<�ڵ����
#define OP_AX_POINT            0x20 ///<ģ���������
#define OP_DX_POINT            0x21 ///<�����������
#define OP_I2_POINT            0x22 ///<�����������
#define OP_I4_POINT            0x23 ///<�����������
#define OP_R8_POINT            0x24 ///<˫���ȵ����

// ���ݿ����
#define OP_SELECT             0 ///< ��ѯ����ֵ
#define OP_UPDATE             1 ///< ��������ֵ
#define OP_INSERT             2 ///< ������ֵ
#define OP_DELETE             3 ///< ɾ������ֵ
#define OP_REPLACE            4 ///< ����о͸��£���û�оͲ���

// �������
#define OPE_OK             0     ///< �ɹ���û�д���
#define OPE_ERROR         -1     ///< δ֪����
#define OPE_PARAM         -10    ///< ��������
#define OPE_UNSUPPORTED   -11    ///< ����δ֧��
#define OPE_MEMORY        -96    ///< �޷������ڴ棬��Ҫ����
#define OPE_NET_IO        -97    ///< �����дIO������Ҫ����
#define OPE_NET_CLOSED    -98    ///< �����ѹرգ���Ҫ����
#define OPE_CONNECT       -99    ///< �޷����ӷ���������Ҫ����
#define OPE_BUFF_NET      -1001  ///< ����Ͽ�
#define OPE_BUFF_IO       -1002  ///< д��洢�����ļ������ˣ����п������ļ����ڷ�������
#define OPE_BUFF_OVERFLOW -1003  ///< �����ļ�����Ԥ��Ĵ�С
/** @} */


typedef int OPDateTime_t;

/* ============================================================================================================================ */
/**
 * @defgroup api2_type_define APIv2 �������Ͷ���
 * @brief ����һЩ������ĳ���������
 * @{
 */

/**
* @brief ͳ��ֵ
*/
typedef struct
{
    int time;        // ʱ���ǩ��ͳ���������ʼʱ��
    short status;       // ͳ�������ڼ�¼�����İٷֱ�0-100
    double flow;        // �ۻ�
    double max;         // ���DX: ����Ϊ1�Ĵ���
    double min;         // ��С��DX: ����Ϊ0�Ĵ���
    int maxtime;       // ���ֵʱ�䣬DX: ֵΪ1����ʱ��
    int mintime;       // ��Сֵʱ�䣬DX: ֵΪ0����ʱ��
    double avg;         // ʱ��ƽ��
    double mean;        // ����ƽ��
    double stdev;       // ��׼����, ��δʵ��
}StatVal;
/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup api2_handle_face APIv2 �����ӽӿ�
 * @brief �������Ӳ��ֽӿڳ�ȡ���������������ӽӿڣ����ǵĹ��ܶ�һ��Ӧ��
 *
 * @details
 *  @li ��һ�麯���� @ref api2_nohandle_face ���ȡ����Ȼ������������ݣ�
 *          �乤��ʵ����������ͬ��
 *  @li �� @ref api2_nohandle_face ���UDP�����;�̬����д����û�г�ȡ����
 *  @li @ref api2_nohandle_face �������ӽӿڵĳ������������ӽӿ��ṩһ���̰߳�ȫ�ĺ�����
 *          ͬʱ����ͳһ����˺���������ʹ�öԺ����Ĺ��ܸ����״Ӻ�������������
 *
 * @{
 */

/**
* @brief ���ݿ���������û�͸��
*/
typedef void* OpenPlant;
/**
* @brief ��������û�͸��
*/
typedef void* OPResult;
/**
* @brief ���ػ��壬�û�͸��
*/
typedef void* OPLocalBuffer;

/**
* @brief openPlant �ͻ��� API v2 ��ʼ��
*
* @param option   ѡ��; �ο���: OP_OPTION_*
* @param host     ������IP������
* @param port     �������˿�
* @param timeout  ���糬ʱ������Ϊ��λ
* @param user     �û���
* @param password ����
* @buffer_path    ��������Ŀ¼���������Ҫ����ָ��Ϊ nullptr
* @buffer_size    �����ļ��ߴ磻�������Ҫ����ָ��Ϊ 0
*
* @return
*   @li �ɹ�: ret != 0������һ�������ַ
*   @li ʧ��: ret == 0
*/
GMacroAPI(OpenPlant) op2_init(int option
        , const char *host , int port , int timeout
        , const char *user , const char *password
        , const char *buffer_path , int buffer_size);

/**
* @brief �ر�����
*
* @param fd ָ������
* @return ������û�з���ֵ
*/
GMacroAPI(void) op2_close(OpenPlant op);

/**
* @brief ��ѯ����������״̬
*
* @param op ���ݿ���������
* @return 0 ��ʾOK��-1 ��ʾ�ѹر�
*/
GMacroAPI(int) op2_status(OpenPlant op);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup apiv2_op_buffer opWriter �������ýӿ�
 * @brief   ��OPBuffer����ֵ��������
 * @details
 *      key��type�����塢[Ĭ��ֵ]
 *  @li transfer_protocol                int     ʹ��tcp����         ��Ĭ��ֵ
 *  @li server_address                   str     ��������ַ          ��Ĭ��ֵ
 *  @li server_port                      int     �������˿ں�        ��Ĭ��ֵ
 *  @li client_name                      str     �ͻ����û���        ��Ĭ��ֵ
 *  @li client_password                  str     �ͻ�������          ��Ĭ��ֵ
 *  @li storage_location                 str     �������ݴ洢·��    ��Ĭ��ֵ
 *  @li storage_capacity                 int     ���������������    ��Ĭ��ֵ
 *  @li option                           int     ��op_writer_open    ��Ĭ��ֵ
 *  @li isolator_enabled                 int     ��������            ��Ĭ��ֵ
 *  @li net_probe_interval               int     ����̽����        5 second
 *  @li upload_history_interval          int     �ϴ���ʷ���        100 ms
 *  @li realtime_filter_enabled          int     �Ƿ����ʵʱֵ      0 bool
 *  @li filter_bool_upload_interval      int     �������ϴ����      30 second
 *  @li filter_bool_storage_interval     int     �������洢���      900 second
 *  @li size_for_each_upload             int     ÿ���ϴ��ļ���С    10 MB
 *  @li upload_history_per_time          int     ÿ���ϴ��ļ�¼��    50000 ��
 *  @li single_file_capacity             int     ÿ�������ļ���С    1024 MB
 *  @li storage_file_quantity            int     ���������ļ�����  ��Ĭ��ֵ
 *  @li storage_time_limit               int     ������������      ��Ĭ��ֵ
 *  @li history_task_interval            int     ��̨�߳�sleepʱ��   1000 ms
 *  @li record_filename                  str     ��¼�ļ���          "op.writer.json"
 *
 * @{
 */


/**
* @brief ��ȡOPBuffer���ڲ�״̬
*
* @param op ���ݿ���
* @param key �ؼ���(key):  network_connection, try_connection_count, cache_status, upload_status
* @param value ��Ӧ���Ե�ֵ
*/
GMacroAPI(int) op2_buffer_get_status(OpenPlant op, const char* key);

/**
* @brief ��OPBuffer����ֵ�������ã�������ֵ����Ϊint
*
* @param op ���ݿ���
* @param name ���Թؼ���(key),��:transfer_protocol (ʹ��tcp����)
* @param value ��Ӧ���Ե�ֵ
*/
GMacroAPI(void) op2_buffer_set_int   (OpenPlant op, const char* name, int value);

/**
* @brief ��OPBuffer����ֵ�������ã�������ֵ����Ϊstring
*
* @param
* @param name ���Թؼ���(key)
* @param value ��Ӧ���Ե�ֵ

*/
GMacroAPI(void) op2_buffer_set_string(OpenPlant op, const char* name, const char* value);

/**
* @brief �Ե��������
* @param op ���ݿ���
* @param id  ���ID
* @param type �������
* @param deadband ����ֵ
*/
GMacroAPI(void) op2_buffer_set_point (OpenPlant op, int id, int type, double deadband);


/** 
 * @brief �򿪱��ػ��壬���ո���ʱ��δ��м�������
 * 
 * @param directroy     �����ļ�����Ŀ¼
 * @param beginTime     Ҫ�������ݵĿ�ʼʱ��
 * @param endTime       Ҫ�������ݵĽ���ʱ��
 * @param error         �����ַ���
 *
 * @return �ɹ�:OPLocalBuffer; ʧ��:NULL
 */
GMacroAPI(OPLocalBuffer) op2_buffer_open(const char* directroy, OPDateTime_t beginTime, OPDateTime_t endTime, const char** error);

/** 
 * @brief ��ȡ�����е���һ��
 * 
 * @param opBuffer op2_buffer_open���صľ��
 *
 * @return ������:����һ�����ݿ�(�����); ������:NULL
 */
GMacroAPI(OPResult) op2_buffer_next(OPLocalBuffer opBuffer);

/** 
 * @brief 
 * 
 * @param result    op2_buffer_next ���ص����ݿ�
 * @param id        ���ݿ�����һ����¼�е�id�ֶ�
 * @param time      ���ݿ�����һ����¼�е�time�ֶ�
 * @param status    ���ݿ�����һ����¼�е�status�ֶ�
 * @param value     ���ݿ�����һ����¼�е�value�ֶ�
 */
GMacroAPI(int) op2_buffer_get_from_result(OPResult result, int *id, int *time, short *status, double *value);

/** 
 * @brief �ͷ�op2_buffer_next ���صĽ����
 * 
 * @param �����ͷŵĽ����
 */
GMacroAPI(void) op2_buffer_free_result(OPResult result);

/** 
 * @brief �ر�op2_buffer_open�򿪵ı��ػ���
 * 
 * @param opBuffer �����رյı��ػ���
 */
GMacroAPI(void) op2_buffer_close(OPLocalBuffer opBuffer);


/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup api2_time_face APIv2 ʱ���������
 * @brief �ѱ���ʱ����ֶ�ʱ�以��
 *
 * @{
 */

/**
* @brief ȡ�����ݿ������ʱ��
*
* @param op ���ݿ���������
* @param out �������ݿ�ʱ��(std::time_t ʱ��ֵ)
* @return ��ȷ ���� 0
*         ���� ���� ����

*/
GMacroAPI(int) op2_get_system_time(OpenPlant op, int*out);

/**
* @brief �� std::time_t ʱ��ֵת��Ϊ �ֶ�ʱ��ֵ���� 2011-11-23 14:14:20
*
* @param t time_t ����ʱ��
* @param yy ���ص���
* @param mm ���ص���
* @param dd ���ص���
* @param hh ���ص�ʱ
* @param mi ���صķ�
* @param ss ���ص���
*
* @return ������������ֵ
*/
GMacroAPI(void) op2_decode_time(int time, int *yy, int *mm, int *dd, int *hh, int *mi, int *ss);

/**
* @brief �� �ֶ�ʱ��ֵ���� 2011-11-23 14:14:20��ת��Ϊ  std::time_t ʱ��ֵ
*
* @param yy �������
* @param mm �������
* @param dd �������
* @param hh �����ʱ
* @param mi ����ķ�
* @param ss �������
*
* @return time_t ����ʱ��
*/
GMacroAPI(int) op2_encode_time(int yy, int mm, int dd, int hh, int mi, int ss);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup point_group APIv2 �����������
 * @brief �������
 * @{
 */

/**
* @brief ��������һ������Ĵ���
* @details
*   @li ��ʱ������Ҫͬʱ��������㣬���Ҵ�ʱֻ֪����������ô�����Ϊ��ʱ׼����
*   @li ����ͨ����������ɡ��ӵ�����ID��ӳ�䣬��ͨ������ȡ��ID
*/
typedef void* OPGroup;

/**
* @brief ����ָ������
* @return ��ȷ: ������������: NULL
*/
GMacroAPI(OPGroup) op2_new_group();

/**
* @brief ȡ��ָ������Ĵ�С
* @param gh ������������ָ��һ������
*/
GMacroAPI(int) op2_group_size(OPGroup gh);

/**
* @brief ��ָ�������������
* @param gh  ���������� @ref op_new_group ����
* @param obj_name ����
*/
GMacroAPI(void) op2_add_group_point(OPGroup gh, const char *obj_name);

/**
* @brief �ͷ�һ������
* @param gh ������
*/
GMacroAPI(void) op2_free_group(OPGroup gh);

/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup api2_result_handle APIv2 �������������
 * @brief �����������ȡֵ���ͷ�
 * @{
 */

/**
* @brief ȡ�ý�����ߴ�
* @param result �����
* @return ������ߴ�
*/
GMacroAPI(int) op2_num_rows(OPResult result);

/**
* @brief ����������һ����ʷֵ�������ȡ��������һ����¼��ֵ
*
* @param result �����
* @param value ���ڷ�����ʷֵ
* @param status ���ڷ���״ֵ̬
* @param time ���ڷ���ʱ���ǩ
*
* @return
*   @li 0: �Ѿ��������ĩβ
*   @li 1: ������һ����¼
*   @li ����: ����
*/
GMacroAPI(int) op2_fetch_timed_value(OPResult result, int* time, short *status, double *value);

/**
* @brief ����������һ��ͳ��ֵ�������ȡ��������һ��ͳ��ֵ
*
* @param result �����
* @param sval ���ڷ��ص�ͳ��ֵ
*
* @return
*   @li 0: �Ѿ��������ĩβ
*   @li 1: ������һ����¼
*   @li ����: ����
*/
GMacroAPI(int) op2_fetch_stat_value(OPResult result, StatVal *sval);

/**
* @brief �ͷ�һ�������
*
* @param result �����
* @return ������û�з���ֵ
*/
GMacroAPI(void) op2_free_result(OPResult result);

/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup op2_get_face APIv2 ȡʵʱ���ݺ���ʷ������ؽӿ�
 * @brief   ͨ����������IDȡʵʱ���ݻ�������ʷ���� 
 *
 * @{
 */

/**
* @brief ͨ������ȡ��ʵʱֵ
*
* @param op ���ݿ���
* @param gh �����ĵ���
* @param time ָ����ʱ�������
* @param status ָ����״̬������
* @param value  ָ����ʵʱֵ������
* @param errors ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ʵʱֵ����
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_value_byname(OpenPlant op, OPGroup gh, int* time, short *status, double *value, int *errors);

/**
* @brief ͨ��IDȡ��ʵʱֵ
*
* @param op  ���ݿ���
* @param num ��ID�ĸ���
* @param id  ָ���ID������
* @param time ָ����ʱ�������
* @param status ָ����״̬������
* @param value  ָ����ʵʱֵ������
* @param errors ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ʵʱֵ����
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_value_byid(OpenPlant op, int num, int *id, int* time, short *status, double *value, int *errors);

/**
* @brief ͨ������ȡ�ÿ���ֵ
*
* @param op ���ݿ���
* @param gh ��������
* @param time ĳһ��ʱ�̵�ֵ��std::time_t)
* @param status ָ����״̬������
* @param value  ָ����Ӧĳһʱ��ֵ������
* @param errors ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_snap_byname(OpenPlant op, OPGroup gh, int time, short *status, double *value, int *errors);

/**
* @brief ͨ����IDȡ�ÿ���ֵ
* @param op ���ݿ���
* @param num ��ID�ĸ���
* @param id  ָ���ID������
* @param time ĳһ��ʱ�̵�ֵ��std::time_t)
* @param status ָ����״̬������
* @param value  ָ����Ӧĳһʱ��ֵ������
* @param errors ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_snap_byid(OpenPlant op, int num, int *id, int time, short *status, double *value, int *errors);

/**
* @brief ͨ������ȡ��ʷ����
*
* @param op ���ݿ���
* @param gh ��������
* @param value_type  ��ʷ��������, �磺ԭʼֵ��0�����ȼ��ֵ��1���ȵ�
* @param begin_tm    ��ʼʱ��
* @param end_tm      ����ʱ��
* @param interval    ���ʱ��
* @param result      ��ʷ�����
* @param errors      ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ֵ����
*
* @return ��ȷ ���� 0
*        ���� ���� ����
*/
GMacroAPI(int) op2_get_history_byname(OpenPlant op, OPGroup gh, int *value_type, int begin_tm, int end_tm, int interval, OPResult *result, int *errors);

/**
* @brief  ͨ����IDȡ��ʷ����
*
* @param op ���ݿ���
* @param num ��ID����
* @param id  ָ���ID������
* @param value_type ��ʷ��������, �磺ԭʼֵ��0�����ȼ��ֵ��1���ȵ�
* @param begin_tm   ��ʼʱ��
* @param end_tm     ����ʱ��
* @param interval   ���ʱ��
* @param result     ��ʷ�����
* @param errors     ָ��ط�����������飬����������Ϊ�����ֵ�������ȡ�õ��ֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_history_byid(OpenPlant op, int num, int *id, int *value_type, int begin_tm, int end_tm, int interval, OPResult *result, int *errors);

/**
* @brief �ѽ�����е�ԭʼֵ��ԭΪ�ȼ��ֵ
*
* @param result ��ʷ���������op2_get_history_byname��op2_get_history_byid�õ�
* @param num    ���ڷ����������� 
* @param time   ���ڷ���ʱ���ǩ���� 
* @param status ���ڷ���״̬���� 
* @param value  ���ڷ��صȼ��ֵ���� 
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_raw_to_span(OPResult result, int *num, int **time, short **status, double **value);

/**
* @brief ͨ��������һ����ȡһ����ʷ���ݣ���ʷ���ݵ�����������
*
* @param op         ���ݿ���������
* @param obj_name    һ������
* @param value_type ��ʷֵ����
* @param begin_tm   ��ʼʱ��
* @param end_tm     ����ʱ��
* @param interval   ʱ����
* @param num        ��ʷ���ݵ�����������num
* @param time     ָ��ʱ���ǩ����
* @param status     ָ��״̬����
* @param value      ָ��ֵ����
* @param actSize    ����ֵʵ�ʵĸ���
*/
GMacroAPI(int) op2_get_histroy_top_byname(OpenPlant op, const char *obj_name, int value_type, int begin_tm, int end_tm, int interval,
                                      int num, int* time, short *status, double *value, int *actSize);

/**
* @brief ͨ����id��һ����ȡһ����ʷ���ݣ���ʷ���ݵ�����������
*
* @param op         ���ݿ���������
* @param id         һ����ID
* @param value_type ��ʷֵ����
* @param begin_tm   ��ʼʱ��
* @param end_tm     ����ʱ��
* @param interval   ʱ����
* @param num        ��ʷ���ݵ�����������num
* @param time     ָ��ʱ���ǩ����
* @param status     ָ��״̬����
* @param value      ָ��ֵ����
* @param actSize    ����ֵʵ�ʵĸ���
*/
GMacroAPI(int) op2_get_histroy_top_byid(OpenPlant op, int id, int value_type, int begin_tm, int end_tm, int interval,
                                    int num, int* time, short *status, double *value, int *actSize);
/** @} */


/* ============================================================================================================================ */
/**
 * @defgroup op2_write_face APIv2 дʵʱ���ݺ���ʷ������ؽӿ�
 * @brief   ͨ����������IDдʵʱ���ݻ�������ʷ���� 
 *
 * @{
 */

/**
* @brief ����дĳһʱ�̵�ʵʱֵ
*
* @param op      ���ݿ���
* @param pt      ������
* @param num     �����
* @param id      ָ���ID����
* @param time  ʱ��ֵ��std:time_t)
* @param status  ָ����״̬������
* @param value   ָ����ֵ������
* @param errors  ָ��ط�����������飬������ֵΪ0����1����ʾд��ɹ�������1��ʾ��һ��д��ɹ���0��ʾ�Ե�ǰʵʱֵ�޸ĳɹ���
*
* @return ��ȷ ���� 0 
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_value(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief ����дĳһʱ�̵�ʵʱֵ(��op2_write_value��������ֻ���дֵ����д״̬)
*
* @param op      ���ݿ���
* @param pt      ������
* @param num     �����
* @param id      ָ���ID����
* @param time  ʱ��ֵ��std:time_t)
* @param status  ָ����״̬������
* @param value   ָ����ֵ������
* @param errors  ָ��ط�����������飬������ֵΪ0����1����ʾд��ɹ�������1��ʾ��һ��д��ɹ���0��ʾ�Ե�ǰʵʱֵ�޸ĳɹ���
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_value_only(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief �����ʵʱֵ(��op2_write_value�������ǣ�ÿ���㶼��Ӧһ��ʱ��)
*
* @param op      ���ݿ���
* @param pt      ������
* @param num     �����
* @param id      ָ���ID����
* @param time  ָ��ʱ��ֵ������
* @param status  ָ����״̬������
* @param value   ָ����ֵ������
* @param errors  ָ��ط�����������飬������ֵΪ0����1����ʾд��ɹ�������1��ʾ��һ��д��ɹ���0��ʾ�Ե�ǰʵʱֵ�޸ĳɹ���
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_value_tm(OpenPlant op, int pt, int num, const int *id, const int* time, const short *status, const double *value, int *errors);

/**
* @brief ʹ�õ�����һ����д�����ʷ����
*
* @param op       ���ݿ���
* @param pt       �������
* @param obj_name  һ������
* @param num      д�����ݵ�����
* @param time   ָ��ʱ���ǩ������
* @param status   ָ���Ӧ��ʱ���ǩ��״̬������
* @param value    ָ�������ʱ���ǩ��ֵ������
* @param error   ָ��ط��������ֵ������ֵ���㣬������õ�дֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_histroy_byname(OpenPlant op, int pt, const char *obj_name, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief ʹ�õ�ID��һ����д�����ʷ����
*
* @param op       ���ݿ���
* @param pt       �������
* @param id       һ�����ID
* @param num      д�����ݵ�����
* @param time   ָ��ʱ���ǩ������
* @param status   ָ���Ӧ��ʱ���ǩ��״̬������
* @param value    ָ�������ʱ���ǩ��ֵ������
* @param error   ָ��ط��������ֵ������ֵ���㣬������õ�дֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_histroy_byid(OpenPlant op, int pt, int id, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief ʹ��ID��������ͬһʱ��д��ʷֵ
*
* @param op       ���ݿ���
* @param pt       ������
* @param num      ��ĸ���
* @param id       ָ���ID������
* @param time   ʱ��ֵ
* @param status   ָ���״̬������
* @param value    ָ���ֵ������
* @param error   ָ��ط��������ֵ������ֵ���㣬������õ�дֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_snap(OpenPlant op, int pt, int num, const int *id, int time, const short *status, const double *value, int *errors);

/**
* @brief ��һ����д������ʷ����
*
* @param op      ���ݿ���
* @param pt      ������
* @param id      ��ID
* @param num     д����ʷ��������
* @param time  ָ���ʱ���ǩ������
* @param status  ָ���Ӧ��ʱ���ǩ��״̬������
* @param value   ָ�������ʱ���ǩ��ֵ������
* @param error   ָ��ط��������ֵ������ֵ���㣬������õ�дֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_cache_one(OpenPlant op, int pt, int id, int num, const int* time, const short *status, const double *value, int *error);

/**
* @brief ͬʱ������д��ʷ����
*
* @param op      ���ݿ���
* @param pt      ������
* @param num     ��ĸ���
* @param id      ָ���ID������
* @param time  ָ���ʱ���ǩ������
* @param status  ָ���Ӧ��ʱ���ǩ��״̬������
* @param value   ָ�������ʱ���ǩ��ֵ������
* @param errors  ָ��ط�����������飬����������Ϊ�����ֵ��������õ�дʵʱֵ����
*
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_write_cache(OpenPlant op, int pt, int num, const int *id, const int* time, const short *status, const double *value, int *errors);

/** @} */

/* ============================================================================================================================ */
/**
 * @defgroup op2_object_handle APIv2 �����ݿ���������صĽӿ�
 * @brief  ����ӿڿ��Զ����ݿ���󣨽ڵ㣬�㣩������ɾ�Ĳ�,����ӵ㣬ɾ�㣬ȡ��̬��Ϣ���޸ľ�̬��Ϣ�ȵ�
 *
 * @{
 */

/**
* @brief ���ݶ����û�͸��
*/
typedef void* OPObject;

/**
* @brief ͨ������ȡ�ö�ӦID
*
* @param op ���ݿ���
* @param gh �����ĵ���
* @param id ָ����ID�����飬���ڴ�ŷ��ص�����Ӧ��ID
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_get_id_byname(OpenPlant op, OPGroup gh, int *id);

/**
* @brief ��������
*
* @param op   ���ݿ���
* @param name Ҫ�������������
* @param type ָ���������ͣ��ο��������ͣ�
* @return
*   @li �ɹ�: ����������
*   @li ʧ��: NULL
*/
GMacroAPI(OPObject) op2_new_object(OpenPlant op, const char *name, int type);

/**
* @brief �ͷŶ����ڴ�ռ�
*
* @param o ������
*/
GMacroAPI(void) op2_free_object(OPObject o);

/**
* @brief ����һ������ľ�̬����ֵ����ֵ����Ϊstring
*
* @param o      ���������֣���op2_new_object���أ�
* @param field  ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @param value  ��̬�����ֶε�ֵ
*/
GMacroAPI(void) op2_object_set_string(OPObject o, const char *field, char *value);

/**
* @brief ����һ������ľ�̬����ֵ����ֵ����Ϊint
*
* @param o      ���������֣���op2_new_object���أ�
* @param field  ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @param value  ��̬�����ֶε�ֵ
*/
GMacroAPI(void) op2_object_set_int   (OPObject o, const char *field, int value);

/**
* @brief ����һ������ľ�̬����ֵ����ֵ����Ϊdouble
*
* @param o      ���������֣���op2_new_object���أ�
* @param field  ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @param value  ��̬�����ֶε�ֵ
*/
GMacroAPI(void) op2_object_set_double(OPObject o, const char *field, double value);

/**
* @brief ȡ��һ������ľ�̬����ֵ����ֵ����Ϊstring
*
* @param o       ���������֣���op2_new_object���أ�
* @param field   ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @param buf     �ַ����飬������ž�̬���Ե�ֵ
* @param len     �ַ�������õ���󳤶�
* @return ��ȷ ���� 0
*         ���� ���� ����
*/
GMacroAPI(int) op2_object_get_string(OPObject o, const char *field, char *buf, int len);

/**
* @brief ȡ��һ������ľ�̬����ֵ����ֵ����Ϊint
*
* @param o       ���������֣���op2_new_object���أ�
* @param field   ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @return  ���� ��̬���Ե�ֵ
*/
GMacroAPI(int) op2_object_get_int   (OPObject o, const char *field);

/**
* @brief ȡ��һ������ľ�̬����ֵ����ֵ����Ϊdouble
*
* @param o       ���������֣���op2_new_object���أ�
* @param field   ��̬���Ե��ֶΣ��磺PN��ED��EU �ȵȣ�
* @return  ���� ��̬���Ե�ֵ
*/
GMacroAPI(double) op2_object_get_double(OPObject o, const char *field);

/**
* @brief ����/ɾ��/���¶���
*
* @param op      ���ݿ���
* @param cmd     �����ݿ�Ĳ��� ��OP_SELECT,OP_UPDATE,OP_INSERT,OP_DELETE,OP_REPLACE)
* @param parent  ������󸸽ڵ�����֣�ȫ����
* @param num     ����ĸ���
* @param objects ��������
* @param errors  �����룬������������0�����ʾ�Ըö���Ĳ���δ�ɹ�
*
* @return
*   @li �ɹ�: ���ز����ɹ��Ķ������
*   @li ʧ��: -1
*/
GMacroAPI(int) op2_modify_object(OpenPlant op, int cmd, const char *parent, int num, OPObject *objects, int *errors);

/**
* @brief ��ȡ���ݶ���
*
* @param op ���ݿ������
* @param gh ������
* @param objects ��������
* @param errors ������
* @return
*   @li �ɹ�: 0
*   @li ʧ��: -1
*/
GMacroAPI(int) op2_get_object_byname(OpenPlant op, OPGroup gh, OPObject *objects, int *errors);

/**
* @brief ��ȡ���ݶ���
*
* @param op ���ݿ������
* @param num �������
* @param num ����ID����
* @param objects ��������
* @param errors ������
* @return
*   @li �ɹ�: 0
*   @li ʧ��: -1
*/
GMacroAPI(int) op2_get_object_byid(OpenPlant op, int num, int *id, OPObject *objects, int *errors);

/**
* @brief ��ȡ���ݿ��б�
*
* @param op ���ݿ������
* @param num          ȡ�������ݿ��б����
* @param databases    ��ȡ�������ݿ�����б�
* @return
*   @li �ɹ�: 0
*   @li ʧ��: ������
*/
GMacroAPI(int) op2_get_database(OpenPlant op, int *num, OPObject **databases);

/**
* @brief ��ȡ�Ӷ����б�
*
* @param op       ���ݿ���������
* @param parent   ָ������������
* @param num      ȡ���Ӷ������
* @param objects  ��ȡ�����Ӷ����б�
* @return
*   @li �ɹ�: 0
*   @li ʧ��: ������
*/
GMacroAPI(int) op2_get_child(OpenPlant op, const char *parent, int *num, OPObject **objects);

/**
* @brief ��ȡ�Ӷ����б�
*
* @param op     ���ݿ���������
* @param parent ָ������������
* @param num    ȡ���Ӷ������
* @param objects ��ȡ�����Ӷ����б�
* @return
*   @li �ɹ�: 0
*   @li ʧ��: ������
*/
GMacroAPI(int) op2_get_child_idname(OpenPlant op, const char *parent, int *num, OPObject **objects);

/**
* @brief �ͷŶ����б�
*
* @param num     �����б�ĸ���
* @param objects �����б�
* @return ������û�з���ֵ
*/
GMacroAPI(void) op2_free_list(int num, OPObject *objects);
/** @} */


#ifdef __cplusplus
}
#endif //__cplusplus
/** @} */

#endif //__OPAPI2_H


