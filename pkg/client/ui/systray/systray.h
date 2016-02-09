#include <wchar.h>

extern void systray_ready();
extern void systray_menu_item_selected(int menu_id);
int nativeLoop(void);
void quit();

#if defined WIN32
void setIcon(const wchar_t* filename, int length);
void setTitle(const wchar_t* title);
void setTooltip(const wchar_t* tooltip);
void add_or_update_menu_item(int menuId, wchar_t* title, wchar_t* tooltip, short disabled, short checked);
#else
void setIcon(const char* iconBytes, int length);
void setTitle(char* title);
void setTooltip(char* tooltip);
void add_or_update_menu_item(int menuId, char* title, char* tooltip, short disabled, short checked);
#endif
